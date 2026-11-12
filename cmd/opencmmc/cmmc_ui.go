package main

// A self-contained, loopback-only browser interface for the offline CMMC workbench.
// It deliberately does not claim certification, authorized assessment, or SPRS submission.
import (
    "crypto/rand"
    "crypto/sha256"
    "crypto/subtle"
    "encoding/hex"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "html/template"
    "io"
    "net"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "runtime"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/blocksifrdev/OpenCMMC/internal/cmmc"
)

const maxUIEvidence = 10 << 20
const maxUICSV = 12 << 20

type localUI struct {
    root, port, secret string
    mu sync.Mutex
}
type localUIState struct {
    Mode string `json:"mode"`
    Level int `json:"level"`
}
type localUIControl struct {
    ID, Description, Status string
}
type localUIPage struct {
    CSRF, Scope, Mode, Selected, SelectedText, SelectedStatus, Summary, Version string
    Controls []localUIControl
    Objectives []cmmc.ObjectiveItem
    Basic *cmmc.Control
    HasWorkspace bool
    IsObjective bool
    MissingCatalog bool
}

func localUISecret() (string, error) {
    b:=make([]byte,32)
    if _,err:=rand.Read(b);err!=nil{return "",err}
    return hex.EncodeToString(b),nil
}
func initUIRoot(root string) (string,error) {
    if root=="" {return "",errors.New("--data-dir required")}
    abs,err:=filepath.Abs(root);if err!=nil{return "",err}
    if err=os.MkdirAll(abs,0700);err!=nil{return "",err}
    for _,dir:=range []string{abs,filepath.Join(abs,"evidence")} {
        if err=os.MkdirAll(dir,0700);err!=nil{return "",err}
        st,e:=os.Lstat(dir);if e!=nil{return "",e}
        if !st.IsDir() || st.Mode()&os.ModeSymlink!=0{return "",fmt.Errorf("unsafe data directory %s",dir)}
        if e=os.Chmod(dir,0700);e!=nil{return "",e}
    }
    return abs,nil
}
func uiWriteJSON(path string, v any) error {
    blob,err:=json.MarshalIndent(v,"","  ");if err!=nil{return err}
    if st,e:=os.Lstat(path);e==nil {
        if !st.Mode().IsRegular(){return errors.New("refusing to replace nonregular workspace")}
    } else if !os.IsNotExist(e){return e}
    temp,err:=os.CreateTemp(filepath.Dir(path),".opencmmc-*.tmp")
    if err!=nil{return err}
    defer os.Remove(temp.Name())
    if err=temp.Chmod(0600);err!=nil{temp.Close();return err}
    if _,err=temp.Write(append(blob,'\n'));err!=nil{temp.Close();return err}
    if err=temp.Sync();err!=nil{temp.Close();return err}
    if err=temp.Close();err!=nil{return err}
    return os.Rename(temp.Name(),path)
}
func uiReadJSON(path string, v any) error {
    st,err:=os.Lstat(path)
    if err!=nil{return err}
    if !st.Mode().IsRegular() || st.Size()>12<<20{return errors.New("invalid workspace file")}
    b,err:=os.ReadFile(path);if err!=nil{return err}
    return json.Unmarshal(b,v)
}
func (u *localUI) state() (localUIState,error) {
    var s localUIState
    err:=uiReadJSON(filepath.Join(u.root,"state.json"),&s)
    if os.IsNotExist(err){return s,nil}
    if err!=nil{return s,err}
    if s.Mode!="basic" && s.Mode!="objectives"{return s,errors.New("invalid UI state")}
    return s,nil
}
func (u *localUI) readBasic() (cmmc.Workspace,error) {
    var w cmmc.Workspace
    err:=uiReadJSON(filepath.Join(u.root,"requirements.json"),&w)
    if err==nil {err=cmmc.ValidatedWorkspace(w)}
    return w,err
}
func (u *localUI) readObjectives() (cmmc.ObjectiveWorkspace,error) {
    var w cmmc.ObjectiveWorkspace
    err:=uiReadJSON(filepath.Join(u.root,"objectives.json"),&w)
    if err==nil {err=cmmc.ValidateObjectiveWorkspace(w)}
    return w,err
}
func localUIHostOK(host,port string) bool {
    h,p,err:=net.SplitHostPort(host)
    return err==nil && h=="127.0.0.1" && (port=="" || p==port)
}
// Browser cookies are scoped by host and path, not TCP port. Distinct
// OpenCMMC instances on 127.0.0.1 must not overwrite each other's sessions.
func (u *localUI) sessionCookieName() string {return "oc_session_p"+u.port}
func (u *localUI) authorized(r *http.Request) bool {
    c,err:=r.Cookie(u.sessionCookieName())
    return err==nil && subtle.ConstantTimeCompare([]byte(c.Value),[]byte(u.secret))==1
}
// formAuthFailure returns a stable diagnostic code, never a secret, Origin
// value, cookie or submitted assessment content. No failed POST executes.
// localUIOriginOK accepts only the local browser origin and the two explicit
// ChromeOS Crostini forwarding aliases documented by Google. ChromeOS may
// rewrite Host to 127.0.0.1 while preserving its browser-visible Origin.
// Never trust generic .test, private-network, wildcard, or remote origins.
func localUIOriginOK(origin,host,port string) bool {
    if !localUIHostOK(host,port) {return false}
    if origin=="" {return true} // allowed by existing clients; CSRF stays mandatory
    parsed,err:=url.Parse(origin)
    if err!=nil || parsed.Scheme!="http" || parsed.User!=nil ||
       parsed.Opaque!="" || parsed.Path!="" || parsed.RawPath!="" ||
       parsed.RawQuery!="" || parsed.Fragment!="" || parsed.Port()!=port {
        return false
    }
    switch parsed.Hostname() {
    case "127.0.0.1","localhost","penguin.linux.test","::1":
        return true
    default:
        return false
    }
}
func (u *localUI) formAuthFailure(r *http.Request) string {
    if r.Method!="POST" || !u.authorized(r) {return "SESSION_INVALID"}
    if !localUIOriginOK(r.Header.Get("Origin"),r.Host,u.port) {
        return "ORIGIN_MISMATCH"
    }
    contentType:=r.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType,"application/x-www-form-urlencoded") &&
       !strings.HasPrefix(contentType,"multipart/form-data") {
        return "UNSUPPORTED_FORM_TYPE"
    }
    if subtle.ConstantTimeCompare([]byte(r.FormValue("csrf")),[]byte(u.secret))!=1 {
        return "FORM_TOKEN_EXPIRED"
    }
    return ""
}
func (u *localUI) csrfOK(r *http.Request) bool {
    return u.formAuthFailure(r)==""
}
func (u *localUI) ServeHTTP(w http.ResponseWriter,r *http.Request) {
    w.Header().Set("X-OpenCMMC-Version",releaseVersion)
    w.Header().Set("Cache-Control","no-store")
    w.Header().Set("Referrer-Policy","no-referrer")
    w.Header().Set("X-Content-Type-Options","nosniff")
    w.Header().Set("X-Frame-Options","DENY")
    w.Header().Set("Content-Security-Policy","default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
    if !localUIHostOK(r.Host,u.port){http.Error(w,"loopback host required",http.StatusForbidden);return}
    if r.URL.Path=="/" && r.Method=="GET" && r.URL.Query().Has("access") {
        if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("access")),[]byte(u.secret))!=1 {
            http.Error(w,"invalid access token",http.StatusForbidden);return
        }
        // A ChromeOS-hosted browser can arrive from the Linux VM through an external
        // top-level navigation. SameSite=Strict may omit the just-issued cookie
        // on the immediate redirect, producing an unauthorized workbench page.
        // Lax permits the safe top-level GET bootstrap while session + exact
        // Origin + CSRF requirements continue to guard all state-changing POSTs.
        http.SetCookie(w,&http.Cookie{Name:u.sessionCookieName(),Value:u.secret,Path:"/",HttpOnly:true,SameSite:http.SameSiteLaxMode})
        http.Redirect(w,r,"/",http.StatusSeeOther);return
    }
    if !u.authorized(r){
        // Do not reveal an access token, lower origin checks, or silently allow
        // unauthenticated requests. Explain how to re-establish a local session.
        w.Header().Set("Content-Type","text/html; charset=utf-8")
        w.WriteHeader(http.StatusForbidden)
        _,_ = io.WriteString(w,`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>OpenCMMC | Reconnect local session</title><style>body{font:16px system-ui,sans-serif;max-width:680px;margin:48px auto;padding:0 18px;line-height:1.55;color:#18283b}main{border:1px solid #d6deea;padding:24px;border-radius:12px;background:#fff}code{background:#edf0f5;padding:3px 5px}</style></head><body><main><h1>Reconnect OpenCMMC</h1><p>This browser tab is not connected to the active local OpenCMMC session. Nothing was changed.</p><ol><li>Return to the terminal running OpenCMMC on this computer.</li><li>Find <strong>Open this private URL on the same computer:</strong> and copy the full address, including the access code.</li><li>Open that exact address in a new tab using <strong>http://127.0.0.1</strong>. Do not substitute localhost, omit the code or use an old bookmark.</li><li>If the terminal no longer shows a running app, close old tabs and restart <code>./start-cmmc-ui.sh</code> from the extracted directory.</li></ol><p>Keep the terminal running while using OpenCMMC. Do not share the private URL or delete your <code>private-assessment</code> folder.</p></main></body></html>`)
        return
    }
    if r.Method=="POST" {
        r.Body=http.MaxBytesReader(w,r.Body,maxUICSV+(2<<20))
        if reason:=u.formAuthFailure(r);reason!="" {
            w.Header().Set("X-OpenCMMC-Form-Diagnostic",reason)
            w.Header().Set("Content-Type","text/html; charset=utf-8")
            if reason=="FORM_TOKEN_EXPIRED" {
                // A new process and an old tab may have different form tokens.
                // Do not replay or execute the rejected POST.
                w.WriteHeader(http.StatusConflict)
                _,_=io.WriteString(w,`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>OpenCMMC | Refresh assessment session</title></head><body><main><h1>Refresh your assessment session</h1><p><strong>Diagnostic: FORM_TOKEN_EXPIRED</strong></p><p>This form is no longer valid. This often happens when you restart OpenCMMC and keep an older browser tab open. The rejected change was not saved.</p><p><a href="/">Reload the current workbench and retry</a></p><p>If the problem repeats, close older tabs and open the newly printed private URL from your OpenCMMC terminal. Never share the access code.</p></main></body></html>`)
                return
            }
            w.WriteHeader(http.StatusForbidden)
            switch reason {
            case "ORIGIN_MISMATCH":
                _,_=io.WriteString(w,`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>OpenCMMC | Browser origin blocked</title></head><body><main><h1>Browser origin blocked</h1><p><strong>Diagnostic: ORIGIN_MISMATCH</strong></p><p>Your browser sent an origin outside OpenCMMC's recognized local addresses. This request was safely rejected; nothing was changed.</p><p>Use the full private address printed in the current terminal. OpenCMMC accepts local HTTP origins for <code>127.0.0.1</code>, <code>localhost</code>, and the ChromeOS Crostini address <code>penguin.linux.test</code> on the same port. Public or HTTPS origins, wrong ports, browser extension pages and opaque origins are still blocked.</p><p>If this happens again, report only this diagnostic code and the address shown in the browser <strong>before any question mark</strong>. Do not share session tokens, cookies, findings or evidence.</p></main></body></html>`)
            case "UNSUPPORTED_FORM_TYPE":
                _,_=io.WriteString(w,`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>OpenCMMC | Form format blocked</title></head><body><main><h1>Unsupported browser form</h1><p><strong>Diagnostic: UNSUPPORTED_FORM_TYPE</strong></p><p>Your browser did not send one of OpenCMMC's supported form formats. The request was not saved. Open the workbench directly on the local URL and use the built-in form, without browser extensions that rewrite requests.</p><p>If this continues, report only this diagnostic code. Never share your private session URL, cookies or evidence.</p></main></body></html>`)
            default:
                _,_=io.WriteString(w,`<!doctype html><html lang="en"><head><meta charset="utf-8"><title>OpenCMMC | Session blocked</title></head><body><main><h1>Session blocked</h1><p><strong>Diagnostic: SESSION_INVALID</strong></p><p>Close old tabs and reopen the full new private URL printed by the OpenCMMC terminal. No data was changed.</p></main></body></html>`)
            }
            return
        }
    }
    u.mu.Lock()
    defer u.mu.Unlock()
    var err error
    switch {
    case r.Method=="GET" && r.URL.Path=="/": err=u.view(w,r)
    case r.Method=="POST" && r.URL.Path=="/create": err=u.create(r)
    case r.Method=="POST" && r.URL.Path=="/import": err=u.importCSV(r)
    case r.Method=="POST" && r.URL.Path=="/decide": err=u.decide(r)
    case r.Method=="POST" && r.URL.Path=="/evidence": err=u.evidence(r)
    case r.Method=="GET" && r.URL.Path=="/export": err=u.export(w,r)
    case r.Method=="GET" && r.URL.Path=="/handoff": err=u.viewHandoff(w,r)
    case r.Method=="POST" && r.URL.Path=="/handoff": err=u.prepareHandoff(w,r)
    default: http.NotFound(w,r);return
    }
    if err!=nil {http.Error(w,"OpenCMMC: "+err.Error(),http.StatusBadRequest);return}
    if r.Method=="POST" && r.URL.Path!="/handoff" {
        dest:="/"
        if control:=r.FormValue("control");control!="" {dest="/?control="+url.QueryEscape(control)}
        http.Redirect(w,r,dest,http.StatusSeeOther)
    }
}
func (u *localUI) create(r *http.Request) error {
    s,err:=u.state();if err!=nil{return err}
    if s.Mode!=""{return errors.New("workspace already exists; use a separate --data-dir for another assessment")}
    level,err:=strconv.Atoi(r.FormValue("level"));if err!=nil{return err}
    scope:=strings.TrimSpace(r.FormValue("scope"))
    if level==2 {
        w,e:=cmmc.NewObjectiveWorkspace(level,scope);if e!=nil{return e}
        if e=uiWriteJSON(filepath.Join(u.root,"objectives.json"),w);e!=nil{return e}
        return uiWriteJSON(filepath.Join(u.root,"state.json"),localUIState{Mode:"objectives",Level:level})
    }
    if level==1 {
        w,e:=cmmc.NewWorkspace(level,scope);if e!=nil{return e}
        if e=uiWriteJSON(filepath.Join(u.root,"requirements.json"),w);e!=nil{return e}
        return uiWriteJSON(filepath.Join(u.root,"state.json"),localUIState{Mode:"basic",Level:level})
    }
    return errors.New("only CMMC Level 1 or Level 2 are supported")
}
func (u *localUI) importCSV(r *http.Request) error {
    state,err:=u.state();if err!=nil{return err}
    if state.Mode!="objectives" {return errors.New("NIST objective import is only supported for Level 2")}
    existing,err:=u.readObjectives();if err!=nil{return err}
    for _,c:=range existing.Controls{for _,o:=range c.Objectives{
        if o.Decision!="UNASSESSED" || len(o.Evidence)>0{return errors.New("cannot replace a workspace containing assessment decisions or evidence")}
    }}
    file,_,err:=r.FormFile("catalog");if err!=nil{return err}
    defer file.Close()
    b,err:=io.ReadAll(io.LimitReader(file,maxUICSV+1));if err!=nil{return err}
    if len(b)>maxUICSV {return errors.New("CSV exceeds 12 MiB")}
    imported,_,err:=cmmc.Import171ACSV(b,existing.Scope,strings.TrimSpace(r.FormValue("expected_sha256")))
    if err!=nil{return err}
    // The actual NIST PDF, not the CSV supplement, is normative.
    return uiWriteJSON(filepath.Join(u.root,"objectives.json"),imported)
}
func (u *localUI) decide(r *http.Request) error {
    state,err:=u.state();if err!=nil{return err}
    id:=r.FormValue("control")
    reviewer:=strings.TrimSpace(r.FormValue("reviewer"))
    rationale:=strings.TrimSpace(r.FormValue("rationale"))
    if len(reviewer)>256 || len(rationale)>2048 {return errors.New("reviewer or rationale too long")}
    if state.Mode=="basic"{
        ws,e:=u.readBasic();if e!=nil{return e}
        for i:=range ws.Controls {
            if ws.Controls[i].ID!=id {continue}
            candidate:=ws.Controls[i]
            candidate.Claim=r.FormValue("decision")
            candidate.Reviewer=reviewer
            candidate.Rationale=rationale
            ws.Controls[i]=candidate
            if e=cmmc.ValidatedWorkspace(ws);e!=nil{return e}
            return uiWriteJSON(filepath.Join(u.root,"requirements.json"),ws)
        }
        return errors.New("unknown requirement")
    }
    if state.Mode=="objectives" {
        ws,e:=u.readObjectives();if e!=nil{return e}
        for ci:=range ws.Controls {
            if ws.Controls[ci].ID!=id{continue}
            for oi:=range ws.Controls[ci].Objectives{
                if ws.Controls[ci].Objectives[oi].ID!=r.FormValue("objective"){continue}
                item:=ws.Controls[ci].Objectives[oi]
                item.Decision=r.FormValue("decision")
                item.Method=r.FormValue("method")
                item.Reviewer=reviewer
                item.Rationale=rationale
                ws.Controls[ci].Objectives[oi]=item
                if e=cmmc.ValidateObjectiveWorkspace(ws);e!=nil{return e}
                return uiWriteJSON(filepath.Join(u.root,"objectives.json"),ws)
            }
            return errors.New("unknown objective")
        }
    }
    return errors.New("no suitable workspace")
}
func (u *localUI) evidence(r *http.Request) error {
    state,err:=u.state();if err!=nil{return err}
    if state.Mode==""{return errors.New("create a workspace first")}
    f,_,err:=r.FormFile("file");if err!=nil{return err}
    defer f.Close()
    base:=filepath.Join(u.root,"evidence")
    st,err:=os.Lstat(base);if err!=nil{return err}
    if !st.IsDir() || st.Mode()&os.ModeSymlink!=0{return errors.New("unsafe evidence directory")}
    temp,err:=os.CreateTemp(base,"evidence-*.bin");if err!=nil{return err}
    keep:=false
    defer func(){if !keep { _=os.Remove(temp.Name()) }}()
    if err=temp.Chmod(0600);err!=nil{temp.Close();return err}
    hasher:=sha256.New()
    n,err:=io.Copy(io.MultiWriter(temp,hasher),io.LimitReader(f,maxUIEvidence+1))
    if err!=nil{temp.Close();return err}
    if n>maxUIEvidence {temp.Close();return errors.New("evidence exceeds 10 MiB")}
    if err=temp.Sync();err!=nil{temp.Close();return err}
    if err=temp.Close();err!=nil{return err}
    ev:=cmmc.Evidence{Path:filepath.Base(temp.Name()),SHA256:hex.EncodeToString(hasher.Sum(nil)),CapturedAt:time.Now().UTC().Format(time.RFC3339)}
    id:=r.FormValue("control")
    if state.Mode=="basic"{
        ws,e:=u.readBasic();if e!=nil{return e}
        for i:=range ws.Controls{if ws.Controls[i].ID==id{
            ws.Controls[i].Evidence=append(ws.Controls[i].Evidence,ev)
            if e=cmmc.ValidatedWorkspace(ws);e!=nil{return e}
            if e=uiWriteJSON(filepath.Join(u.root,"requirements.json"),ws);e!=nil{return e}
            keep=true
            return nil
        }}
    }else {
        ws,e:=u.readObjectives();if e!=nil{return e}
        for ci:=range ws.Controls{if ws.Controls[ci].ID==id{
            for oi:=range ws.Controls[ci].Objectives{
                if ws.Controls[ci].Objectives[oi].ID==r.FormValue("objective"){
                    ws.Controls[ci].Objectives[oi].Evidence=append(ws.Controls[ci].Objectives[oi].Evidence,cmmc.ObjectiveEvidence{Path:ev.Path,SHA256:ev.SHA256,CapturedAt:ev.CapturedAt})
                    if e=cmmc.ValidateObjectiveWorkspace(ws);e!=nil{return e}
                    if e=uiWriteJSON(filepath.Join(u.root,"objectives.json"),ws);e!=nil{return e}
                    keep=true
                    return nil
                }
            }
        }}
    }
    return errors.New("unknown evidence destination")
}
func (u *localUI) assessed() (any,map[string]string,string,error){
    state,err:=u.state();if err!=nil{return nil,nil,"",err}
    statuses:=map[string]string{}
    if state.Mode=="basic"{
        w,e:=u.readBasic();if e!=nil{return nil,nil,"",e}
        report,e:=cmmc.Evaluate(w,filepath.Join(u.root,"evidence"),time.Now().UTC(),90);if e!=nil{return nil,nil,"",e}
        for _,f:=range report.Findings{statuses[f.ID]=f.Status}
        return report,statuses,fmt.Sprintf("Documented: %d | Gaps: %d | Unassessed: %d | Unverifiable: %d | Scope review: %d",report.Summary.Documented,report.Summary.Gaps,report.Summary.Unassessed,report.Summary.Unverifiable,report.Summary.ScopeReview),nil
    }
    if state.Mode=="objectives" {
        w,e:=u.readObjectives();if e!=nil{return nil,nil,"",e}
        report,e:=cmmc.AssessObjectives(w,filepath.Join(u.root,"evidence"),time.Now().UTC(),90);if e!=nil{return nil,nil,"",e}
        for _,f:=range report.Findings{statuses[f.ControlID]=f.Status}
        return report,statuses,fmt.Sprintf("Review-ready evidence: %d | Reported gaps: %d | Unresolved: %d | Empty objective lists: %d",report.Summary.ReviewReady,report.Summary.GapsReported,report.Summary.Unresolved,report.Summary.MissingObjectiveLists),nil
    }
    return nil,nil,"",errors.New("no workspace created")
}
func (u *localUI) export(w http.ResponseWriter,r *http.Request) error {
    result,_,_,err:=u.assessed();if err!=nil{return err}
    b,err:=json.MarshalIndent(result,"","  ");if err!=nil{return err}
    w.Header().Set("Content-Type","application/json; charset=utf-8")
    w.Header().Set("Content-Disposition",`attachment; filename="cmmc-readiness.json"`)
    _,err=w.Write(append(b,'\n'));return err
}
func (u *localUI) view(w http.ResponseWriter,r *http.Request) error {
    state,err:=u.state();if err!=nil{return err}
    p:=localUIPage{CSRF:u.secret,Version:releaseVersion,Mode:state.Mode,HasWorkspace:state.Mode!=""}
    if !p.HasWorkspace{return uiHTML.Execute(w,p)}
    _,statuses,summary,err:=u.assessed();if err!=nil{return err}
    p.Summary=summary
    sel:=r.URL.Query().Get("control")
    if state.Mode=="basic"{
        ws,e:=u.readBasic();if e!=nil{return e}
        p.Scope=ws.Scope
        for i:=range ws.Controls {
            c:=ws.Controls[i]
            p.Controls=append(p.Controls,localUIControl{ID:c.ID,Description:c.Family,Status:statuses[c.ID]})
            if sel==c.ID || sel=="" && i==0 {
                p.Selected=c.ID;p.SelectedText=c.Family;p.SelectedStatus=statuses[c.ID]
                copy:=c;p.Basic=&copy
            }
        }
    }else{
        ws,e:=u.readObjectives();if e!=nil{return e}
        p.Scope=ws.Scope;p.IsObjective=true
        for i:=range ws.Controls{
            c:=ws.Controls[i]
            p.Controls=append(p.Controls,localUIControl{ID:c.ID,Description:c.RequirementText,Status:statuses[c.ID]})
            if sel==c.ID || sel=="" && i==0{
                p.Selected=c.ID;p.SelectedText=c.RequirementText;p.SelectedStatus=statuses[c.ID]
                p.Objectives=c.Objectives
                p.MissingCatalog=len(c.Objectives)==0
            }
        }
    }
    w.Header().Set("Content-Type","text/html; charset=utf-8")
    return uiHTML.Execute(w,p)
}
func openLocalBrowser(target string) error {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        cmd=exec.Command("rundll32","url.dll,FileProtocolHandler",target)
    case "darwin":
        cmd=exec.Command("open",target)
    default:
        cmd=exec.Command("xdg-open",target)
    }
    if err:=cmd.Start();err!=nil{return err}
    go func(){_ = cmd.Wait()}()
    return nil
}
func cmmcUICommand(args []string,out io.Writer) error {
    fs:=flag.NewFlagSet("cmmc ui",flag.ContinueOnError)
    fs.SetOutput(io.Discard)
    dir:=fs.String("data-dir","./opencmmc-workspace","private local assessment directory")
    port:=fs.Int("port",8765,"127.0.0.1-only port (1024..65535)")
    autoOpen:=fs.Bool("open-browser",false,"open the private loopback URL in the default browser")
    if err:=fs.Parse(args);err!=nil{return err}
    if fs.NArg()!=0{return errors.New("unexpected positional arguments")}
    if *port<1024 || *port>65535{return errors.New("port must be 1024..65535")}
    root,err:=initUIRoot(*dir);if err!=nil{return err}
    token,err:=localUISecret();if err!=nil{return err}
    ln,err:=net.Listen("tcp4",net.JoinHostPort("127.0.0.1",strconv.Itoa(*port)));if err!=nil{return err}
    defer ln.Close()
    handler:=&localUI{root:root,port:strconv.Itoa(*port),secret:token}
    server:=&http.Server{Handler:handler,ReadHeaderTimeout:5*time.Second,ReadTimeout:45*time.Second,WriteTimeout:60*time.Second,IdleTimeout:40*time.Second,MaxHeaderBytes:8192}
    fmt.Fprintln(out,"OpenCMMC",releaseVersion,"LOCAL ONLY | no cloud, telemetry, certification or government submission.")
    privateURL:=fmt.Sprintf("http://127.0.0.1:%d/?access=%s",*port,token)
    fmt.Fprintln(out,"Open this private URL on the same computer:",privateURL)
    fmt.Fprintln(out,"Keep the URL private. Press Ctrl+C to stop. Evidence remains in",root)
    if *autoOpen {
        if err:=openLocalBrowser(privateURL);err!=nil {
            fmt.Fprintln(out,"Could not launch the browser automatically; open the private URL manually:",err)
        }
    }
    return server.Serve(ln)
}

// No external scripts, remote fonts, CDN dependencies, or browser network requests.
var uiHTML=template.Must(template.New("ui").Parse(`<!doctype html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>OpenCMMC | Offline CMMC workbench</title>
<style>
:root{font:16px system-ui,sans-serif;color:#1e2c3c;background:#f5f7fb}*{box-sizing:border-box}
body{margin:0}.top{background:#102440;color:white;padding:22px 30px}.top h1{margin:4px 0;font-size:24px}
main{max-width:1300px;margin:auto;padding:20px}.intro,.panel{background:white;border:1px solid #dce3ec;border-radius:12px;padding:20px;margin-bottom:15px}
a{color:#13579c}button{border:0;background:#1554a0;color:white;border-radius:7px;padding:11px 15px;cursor:pointer}
input,select,textarea{font:inherit;border:1px solid #abb9c7;border-radius:6px;padding:9px;width:100%;max-width:550px}
textarea{min-height:72px}.muted{color:#506075;font-size:14px}.grid{display:grid;grid-template-columns:minmax(230px,310px) 1fr;gap:18px}
.sidebar{max-height:70vh;overflow:auto}.control{display:block;padding:9px 6px;border-bottom:1px solid #e4e9ef;font-size:13px}
label{display:block;margin-top:12px;font-weight:600}form{padding:15px 0;border-top:1px solid #e1e7ef}
.item{padding:16px;border:1px solid #e2e7ee;border-radius:9px;margin:12px 0}small{color:#4d5c6e}
.optional-services{border-color:#cfc7b8;background:#fffdfa}.service-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}.service-card{padding:15px;border:1px solid #ded7cd;border-radius:10px;background:#fff}.service-card h3{margin:0;font-size:17px}.service-card p{font-size:14px;line-height:1.5}.service-card a{font-weight:650}@media(max-width:900px){.service-grid{grid-template-columns:1fr}}
@media(max-width:740px){.grid{grid-template-columns:1fr}.sidebar{max-height:250px}}
</style></head><body><header class="top"><div>BLOCKSIFR / OPEN SOURCE</div><h1>OpenCMMC</h1><div>Local CMMC evidence workbench | No subscription, no cloud</div></header>
<main><div class="intro"><small>OpenCMMC build: <strong>{{.Version}}</strong></small><p><strong>Readiness, not certification.</strong> Reviewer decisions and matching hashes do not establish security control effectiveness. No SPRS submission, C3PAO accreditation or government endorsement. Keep CUI inside an approved boundary.</p></div>
{{if not .HasWorkspace}}
<section class="panel"><h2>Create an offline assessment</h2><p>Choose your baseline and name the actual system or enclave boundary. Your data will stay on this device.</p>
<form method="post" action="/create"><input type="hidden" name="csrf" value="{{.CSRF}}">
<label>CMMC level<select name="level"><option value="2">Level 2, 110 requirements</option><option value="1">Level 1, 15 safeguarding items</option></select></label>
<label>System / enclave scope<input name="scope" maxlength="256" required placeholder="Example: Nonproduction demo enclave"></label>
<p><button type="submit">Create assessment</button></p></form></section>
{{else}}
<section class="panel"><h2>{{.Scope}}</h2><p class="muted">{{.Summary}}</p>
<p><a href="/export">Export local JSON findings</a> | <small>Results must be independently reviewed before external use.</small></p>
{{if .IsObjective}}
<details><summary><strong>Import NIST assessment objectives from the official CSV (Level 2)</strong></summary>
<p class="muted">Obtain the official NIST SP 800-171A supplemental CSV and expected digest independently, using an approved staging process. This importer does not reconcile the CSV with the authoritative PDF. Import is blocked after decisions have been recorded.</p>
<form method="post" action="/import" enctype="multipart/form-data"><input type="hidden" name="csrf" value="{{.CSRF}}">
<label>Official NIST CSV<input type="file" name="catalog" accept=".csv,text/csv" required></label>
<label>Independently verified SHA-256<input name="expected_sha256" required pattern="[0-9a-f]{64}" maxlength="64"></label>
<p><button type="submit">Import objective catalog</button></p></form></details>
{{end}}</section>
<div class="grid"><aside class="panel sidebar"><h2>Requirements</h2>{{range .Controls}}<a class="control" href="/?control={{.ID}}"><strong>{{.ID}}</strong> | {{.Status}}</a>{{end}}</aside>
<section class="panel"><h2>Requirement {{.Selected}}</h2><p>{{.SelectedText}}</p><p><strong>Status: {{.SelectedStatus}}</strong></p>
{{if .IsObjective}}
{{if .MissingCatalog}}<p>Objective list missing. Import the official version-pinned CSV to work at objective level. This requirement is not established.</p>{{end}}
{{range .Objectives}}
<div class="item"><h3>{{.ID}}</h3><p>{{.Statement}}</p><p>Current reviewer decision: <strong>{{.Decision}}</strong></p>
<form method="post" action="/decide"><input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="control" value="{{$.Selected}}"><input type="hidden" name="objective" value="{{.ID}}">
<label>Assessment decision<select name="decision"><option value="UNASSESSED" {{if eq .Decision "UNASSESSED"}}selected{{end}}>UNASSESSED</option><option value="ASSERTED_MET" {{if eq .Decision "ASSERTED_MET"}}selected{{end}}>Reviewer asserts met</option><option value="ASSERTED_NOT_MET" {{if eq .Decision "ASSERTED_NOT_MET"}}selected{{end}}>Reviewer reports gap</option><option value="NOT_APPLICABLE" {{if eq .Decision "NOT_APPLICABLE"}}selected{{end}}>Applicability review required</option></select></label>
<label>Method<select name="method"><option value="">Select method</option><option value="EXAMINE" {{if eq .Method "EXAMINE"}}selected{{end}}>EXAMINE</option><option value="INTERVIEW" {{if eq .Method "INTERVIEW"}}selected{{end}}>INTERVIEW</option><option value="TEST" {{if eq .Method "TEST"}}selected{{end}}>TEST</option></select></label>
<label>Reviewer<input name="reviewer" value="{{.Reviewer}}" maxlength="256"></label>
<label>Reason / test observation<textarea name="rationale" maxlength="2048">{{.Rationale}}</textarea></label>
<p><button type="submit">Save objective decision</button></p></form>
<form method="post" action="/evidence" enctype="multipart/form-data"><input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="control" value="{{$.Selected}}"><input type="hidden" name="objective" value="{{.ID}}">
<label>Attach local evidence file (up to 10 MiB)<input type="file" name="file" required></label><p><button type="submit">Attach evidence locally</button></p></form>
<p class="muted">{{len .Evidence}} local evidence reference(s). Attachment integrity alone does not prove effectiveness.</p></div>
{{end}}
{{else}}
{{with .Basic}}
<form method="post" action="/decide"><input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="control" value="{{$.Selected}}">
<label>Reviewer assertion<select name="decision"><option value="UNASSESSED" {{if eq .Claim "UNASSESSED"}}selected{{end}}>UNASSESSED</option><option value="IMPLEMENTED" {{if eq .Claim "IMPLEMENTED"}}selected{{end}}>Reviewer asserts implemented</option><option value="GAP" {{if eq .Claim "GAP"}}selected{{end}}>Reported gap</option><option value="NOT_APPLICABLE" {{if eq .Claim "NOT_APPLICABLE"}}selected{{end}}>Scope review required</option></select></label>
<label>Reviewer<input name="reviewer" value="{{.Reviewer}}" maxlength="256"></label>
<label>Reason / observation<textarea name="rationale" maxlength="2048">{{.Rationale}}</textarea></label>
<p><button type="submit">Save requirement review</button></p></form>
<form method="post" action="/evidence" enctype="multipart/form-data"><input type="hidden" name="csrf" value="{{$.CSRF}}"><input type="hidden" name="control" value="{{$.Selected}}">
<label>Attach local evidence file (up to 10 MiB)<input type="file" name="file" required></label><p><button type="submit">Attach evidence locally</button></p></form>
<p class="muted">{{len .Evidence}} local evidence reference(s).</p>
{{end}}
{{end}}
</section></div>
<section class="panel optional-services" aria-label="Optional BlockSiFr services">
<h2>Want to go further?</h2>
<p>OpenCMMC's local readiness assessment remains free. These are <strong>optional, separately operated BlockSiFr services</strong>, not required to complete your report.</p>
<div class="service-grid">
<article class="service-card"><h3>CortexTrace</h3><p><strong>Fixing findings with an AI coding assistant?</strong> Explore governed coding sessions, change-impact review and execution evidence before merge or deployment.</p><p><a href="/handoff?service=cortextrace">Prepare a private CortexTrace request</a> | <a href="https://exchange.blocksifr.com/solutions/software-delivery-governance" target="_blank" rel="noopener noreferrer">Service overview</a></p></article>
<article class="service-card"><h3>BlockSiFr Control</h3><p><strong>Should agents be allowed to make changes?</strong> Explore runtime authority, constraints, human step-up decisions and execution receipts for consequential actions.</p><p><a href="/handoff?service=control">Prepare a private Control request</a> | <a href="https://www.blocksifr.com/control" target="_blank" rel="noopener noreferrer">Service overview</a></p></article>
<article class="service-card"><h3>BLAKE</h3><p><strong>Could there be more high-impact vulnerabilities?</strong> Explore deeper, consequence-focused security discovery and validation beyond the OpenCMMC readiness review.</p><p><a href="/handoff?service=blake">Prepare a private BLAKE request</a> | <a href="https://www.blocksifr.com/" target="_blank" rel="noopener noreferrer">Service overview</a></p></article>
</div>
<p class="muted"><strong>Your data stays here.</strong> No workspace, control identifier, evidence, file path or report is sent to these sites by OpenCMMC. Links open external websites <em>only when you choose them</em> and require network access. Before providing any CUI, sensitive files or repository access to a service, obtain your organization's authorization. These links do not start a coding session, enforce a policy, or run a vulnerability scan.</p>
</section>{{end}}
<footer class="muted">OpenCMMC by BlockSiFr | Offline-first | Not an approved assessment methodology or government certification</footer></main></body></html>`))

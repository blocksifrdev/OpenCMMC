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

    "github.com/blocksifrdev/freecompliance/internal/cmmc"
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
    CSRF, Scope, Mode, Selected, SelectedText, SelectedStatus, Summary string
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
    temp,err:=os.CreateTemp(filepath.Dir(path),".freecompliance-*.tmp")
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
func (u *localUI) authorized(r *http.Request) bool {
    c,err:=r.Cookie("fc_session")
    return err==nil && subtle.ConstantTimeCompare([]byte(c.Value),[]byte(u.secret))==1
}
func (u *localUI) csrfOK(r *http.Request) bool {
    if !u.authorized(r) || r.Method!="POST" {return false}
    if origin:=r.Header.Get("Origin");origin!="" && origin!="http://"+r.Host{return false}
    if !strings.HasPrefix(r.Header.Get("Content-Type"),"application/x-www-form-urlencoded") &&
       !strings.HasPrefix(r.Header.Get("Content-Type"),"multipart/form-data"){return false}
    return subtle.ConstantTimeCompare([]byte(r.FormValue("csrf")),[]byte(u.secret))==1
}
func (u *localUI) ServeHTTP(w http.ResponseWriter,r *http.Request) {
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
        http.SetCookie(w,&http.Cookie{Name:"fc_session",Value:u.secret,Path:"/",HttpOnly:true,SameSite:http.SameSiteStrictMode})
        http.Redirect(w,r,"/",http.StatusSeeOther);return
    }
    if !u.authorized(r){http.Error(w,"Open the private local URL shown by opencmmc cmmc ui",http.StatusForbidden);return}
    if r.Method=="POST" {
        r.Body=http.MaxBytesReader(w,r.Body,maxUICSV+(2<<20))
        if !u.csrfOK(r){
        // A browser tab left open across a process restart can hold an old
        // form token while the browser already has the replacement session
        // cookie. Never execute such a POST: offer a safe, one-click refresh.
        origin:=r.Header.Get("Origin")
        contentType:=r.Header.Get("Content-Type")
        sameOrigin:=origin=="" || origin=="http://"+r.Host
        supportedForm:=strings.HasPrefix(contentType,"application/x-www-form-urlencoded") ||
            strings.HasPrefix(contentType,"multipart/form-data")
        if u.authorized(r) && sameOrigin && supportedForm {
            w.Header().Set("Content-Type","text/html; charset=utf-8")
            w.WriteHeader(http.StatusConflict)
            _,_ = io.WriteString(w,`<!doctype html><html lang="en"><head><meta charset="utf-8"><title>OpenCMMC session refresh</title></head><body><main><h1>Refresh your assessment session</h1><p>This form is no longer current. This often happens when a browser tab stays open after the OpenCMMC application is restarted. No assessment changes were saved.</p><p><a href="/">Reload the current workbench and retry</a></p><p>If reloading does not help, close older tabs and open the newly printed private URL from your OpenCMMC command window. Keep that URL private.</p></main></body></html>`)
            return
        }
        http.Error(w,"invalid form token or request origin",http.StatusForbidden)
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
    default: http.NotFound(w,r);return
    }
    if err!=nil {http.Error(w,"OpenCMMC: "+err.Error(),http.StatusBadRequest);return}
    if r.Method=="POST" {
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
    p:=localUIPage{CSRF:u.secret,Mode:state.Mode,HasWorkspace:state.Mode!=""}
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
    dir:=fs.String("data-dir","./freecompliance-workspace","private local assessment directory")
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
    fmt.Fprintln(out,"OpenCMMC LOCAL ONLY | no cloud, telemetry, certification or government submission.")
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
@media(max-width:740px){.grid{grid-template-columns:1fr}.sidebar{max-height:250px}}
</style></head><body><header class="top"><div>BLOCKSIFR / OPEN SOURCE</div><h1>OpenCMMC</h1><div>Local CMMC evidence workbench | No subscription, no cloud</div></header>
<main><div class="intro"><strong>Readiness, not certification.</strong> Reviewer decisions and matching hashes do not establish security control effectiveness. No SPRS submission, C3PAO accreditation or government endorsement. Keep CUI inside an approved boundary.</div>
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
</section></div>{{end}}
<footer class="muted">OpenCMMC by BlockSiFr | Offline-first | Not an approved assessment methodology or government certification</footer></main></body></html>`))

package main

import (
    "bytes"
    "encoding/json"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/blocksifrdev/OpenCMMC/internal/cmmc"
)

func testUI(t *testing.T) *localUI {
    t.Helper()
    root,err:=initUIRoot(t.TempDir())
    if err!=nil{t.Fatal(err)}
    return &localUI{root:root,port:"8765",secret:strings.Repeat("a",64)}
}
func uiRequest(t *testing.T,u *localUI,method,path string,form url.Values,authenticated bool,origin string) *httptest.ResponseRecorder {
    t.Helper()
    var req *http.Request
    if method=="POST" {
        if form==nil{form=url.Values{}}
        req=httptest.NewRequest(method,"http://127.0.0.1:8765"+path,strings.NewReader(form.Encode()))
        req.Header.Set("Content-Type","application/x-www-form-urlencoded")
    }else {req=httptest.NewRequest(method,"http://127.0.0.1:8765"+path,nil)}
    if authenticated {req.AddCookie(&http.Cookie{Name:u.sessionCookieName(),Value:u.secret})}
    if origin!=""{req.Header.Set("Origin",origin)}
    response:=httptest.NewRecorder()
    u.ServeHTTP(response,req)
    return response
}
func TestUILoopbackAuthenticationAndCSRF(t *testing.T){
    u:=testUI(t)
    if r:=uiRequest(t,u,"GET","/",nil,false,"");r.Code!=http.StatusForbidden{t.Fatalf("unauthenticated request: %d",r.Code)}
    if r:=uiRequest(t,u,"GET","/?access=bad",nil,false,"");r.Code!=http.StatusForbidden{t.Fatalf("bad token: %d",r.Code)}
    if r:=uiRequest(t,u,"GET","/?access="+u.secret,nil,false,"");r.Code!=http.StatusSeeOther||len(r.Result().Cookies())==0{t.Fatalf("login failed: %d",r.Code)}
    if r:=uiRequest(t,u,"POST","/create",url.Values{"scope":{"Demo"},"level":{"1"},"csrf":{"bad"}},true,"");r.Code!=http.StatusConflict{t.Fatalf("bad csrf must not execute and must prompt refresh: %d",r.Code)}
    if r:=uiRequest(t,u,"POST","/create",url.Values{"scope":{"Demo"},"level":{"1"},"csrf":{u.secret}},true,"https://evil.example");r.Code!=http.StatusForbidden{t.Fatalf("cross-origin request: %d",r.Code)}
    r:=httptest.NewRecorder()
    badHost:=httptest.NewRequest("GET","http://evil.example:8765/",nil)
    badHost.AddCookie(&http.Cookie{Name:u.sessionCookieName(),Value:u.secret})
    u.ServeHTTP(r,badHost)
    if r.Code!=http.StatusForbidden{t.Fatalf("host rebinding accepted: %d",r.Code)}
    if localUIHostOK("0.0.0.0:8765","8765")||localUIHostOK("127.0.0.1:8766","8765"){t.Fatal("unsafe binding accepted")}
}
func TestUIStaleFormTokenRecoveryWithoutStateChange(t *testing.T){
    u:=testUI(t)
    stale:=uiRequest(t,u,"POST","/create",url.Values{
        "scope":{"Synthetic ARM64 assessment"},
        "level":{"1"},
        "csrf":{"a-token-from-an-old-tab"},
    },true,"http://127.0.0.1:8765")
    if stale.Code!=http.StatusConflict {
        t.Fatalf("expired form should prompt refresh: %d %s",stale.Code,stale.Body.String())
    }
    if !strings.Contains(stale.Body.String(),"Reload the current workbench") {
        t.Fatal("expired form is missing recovery link")
    }
    if _,err:=os.Stat(filepath.Join(u.root,"state.json"));!os.IsNotExist(err) {
        t.Fatalf("stale request must not create a workspace: %v",err)
    }
    refreshed:=uiRequest(t,u,"GET","/",nil,true,"")
    if refreshed.Code!=http.StatusOK || !strings.Contains(refreshed.Body.String(),u.secret) {
        t.Fatal("reloaded page must supply the current form token")
    }
    accepted:=uiRequest(t,u,"POST","/create",url.Values{
        "scope":{"Synthetic ARM64 assessment"},
        "level":{"1"},
        "csrf":{u.secret},
    },true,"http://127.0.0.1:8765")
    if accepted.Code!=http.StatusSeeOther {
        t.Fatalf("fresh form should succeed: %d %s",accepted.Code,accepted.Body.String())
    }
}

func TestUILocalLevel1ReviewAndExport(t *testing.T){
    u:=testUI(t)
    create:=uiRequest(t,u,"POST","/create",url.Values{"scope":{"Synthetic level 1 test"},"level":{"1"},"csrf":{u.secret}},true,"http://127.0.0.1:8765")
    if create.Code!=http.StatusSeeOther{t.Fatalf("create: %d %s",create.Code,create.Body.String())}
    ws,err:=u.readBasic();if err!=nil{t.Fatal(err)}
    if len(ws.Controls)!=15{t.Fatalf("want 15 controls, got %d",len(ws.Controls))}
    id:=ws.Controls[0].ID
    gap:=uiRequest(t,u,"POST","/decide",url.Values{"csrf":{u.secret},"control":{id},"decision":{"GAP"},"reviewer":{"Tester"},"rationale":{"Synthetic gap"}},true,"")
    if gap.Code!=http.StatusSeeOther{t.Fatalf("review: %d %s",gap.Code,gap.Body.String())}
    export:=uiRequest(t,u,"GET","/export",nil,true,"")
    if export.Code!=http.StatusOK{t.Fatalf("export: %d %s",export.Code,export.Body.String())}
    var report cmmc.Report
    if err=json.Unmarshal(export.Body.Bytes(),&report);err!=nil{t.Fatal(err)}
    if report.Summary.Gaps!=1 || report.Summary.Unassessed!=14{t.Fatalf("unexpected summary: %+v",report.Summary)}
    if report.OfficialSubmission!="NOT_PERFORMED"{t.Fatal("falsely claimed submission")}
    if !strings.Contains(export.Header().Get("Content-Disposition"),"attachment;"){t.Fatal("report should download")}
}
func TestUIEvidenceUploadStoredLocally(t *testing.T){
    u:=testUI(t)
    createReq:=httptest.NewRequest("POST","http://127.0.0.1:8765/create",strings.NewReader(url.Values{"level":{"1"},"scope":{"Synthetic demo"}}.Encode()))
    createReq.Header.Set("Content-Type","application/x-www-form-urlencoded")
    if err:=u.create(createReq);err!=nil{t.Fatal(err)}
    ws,err:=u.readBasic();if err!=nil{t.Fatal(err)}
    id:=ws.Controls[0].ID
    var body bytes.Buffer
    multi:=multipart.NewWriter(&body)
    for k,v:=range map[string]string{"csrf":u.secret,"control":id} {if err=multi.WriteField(k,v);err!=nil{t.Fatal(err)}}
    file,err:=multi.CreateFormFile("file","demo.txt");if err!=nil{t.Fatal(err)}
    if _,err=file.Write([]byte("synthetic evidence only"));err!=nil{t.Fatal(err)}
    if err=multi.Close();err!=nil{t.Fatal(err)}
    req:=httptest.NewRequest("POST","http://127.0.0.1:8765/evidence",&body)
    req.Header.Set("Content-Type",multi.FormDataContentType())
    req.AddCookie(&http.Cookie{Name:u.sessionCookieName(),Value:u.secret})
    result:=httptest.NewRecorder();u.ServeHTTP(result,req)
    if result.Code!=http.StatusSeeOther{t.Fatalf("upload: %d %s",result.Code,result.Body.String())}
    ws,err=u.readBasic();if err!=nil{t.Fatal(err)}
    if len(ws.Controls[0].Evidence)!=1{t.Fatal("no attached evidence")}
    evidence:=ws.Controls[0].Evidence[0]
    stored,err:=os.ReadFile(filepath.Join(u.root,"evidence",evidence.Path))
    if err!=nil{t.Fatal(err)}
    if string(stored)!="synthetic evidence only"{t.Fatal("evidence was changed")}
    result=uiRequest(t,u,"POST","/decide",url.Values{"csrf":{u.secret},"control":{id},"decision":{"IMPLEMENTED"},"reviewer":{"test reviewer"}},true,"")
    if result.Code!=http.StatusSeeOther{t.Fatalf("decision: %d %s",result.Code,result.Body.String())}
    output,_,_,err:=u.assessed();if err!=nil{t.Fatal(err)}
    if output.(cmmc.Report).Summary.Documented!=1 {t.Fatal("expected one evidence-integrity DOCUMENTED assertion")}
}
func TestUIEmptyObjectiveCatalogCannotPass(t *testing.T){
    u:=testUI(t)
    req:=httptest.NewRequest("POST","http://127.0.0.1:8765/create",strings.NewReader(url.Values{"level":{"2"},"scope":{"Synthetic level 2 test"}}.Encode()))
    req.Header.Set("Content-Type","application/x-www-form-urlencoded")
    if err:=u.create(req);err!=nil{t.Fatal(err)}
    w,err:=u.readObjectives();if err!=nil{t.Fatal(err)}
    if len(w.Controls)!=110{t.Fatal("missing Level 2 identifiers")}
    result,_,_,err:=u.assessed();if err!=nil{t.Fatal(err)}
    r:=result.(cmmc.ObjectiveReport)
    if r.Summary.MissingObjectiveLists!=110 || r.Summary.ReviewReady!=0{t.Fatalf("unexpected assessment: %+v",r.Summary)}
    if r.SPRSScore!="NOT_CALCULATED"{t.Fatal("incorrect SPRS status")}
}

// TestUIBrowserBootstrapRecoveryOnChromeOS checks browser behavior at the
// application boundary. SameSite=Lax prevents a newly set session cookie from
// being suppressed by an external app-to-Chrome first navigation and redirect.
func TestUIBrowserBootstrapRecoveryOnChromeOS(t *testing.T) {
    u:=testUI(t)
    first:=uiRequest(t,u,"GET","/",nil,false,"")
    if first.Code!=http.StatusForbidden {
        t.Fatalf("missing session should explain recovery, got %d",first.Code)
    }
    for _,hint:=range []string{"Reconnect OpenCMMC","private URL","127.0.0.1","start-cmmc-ui.sh"} {
        if !strings.Contains(first.Body.String(),hint) {t.Errorf("missing recovery instruction %q",hint)}
    }
    if strings.Contains(first.Body.String(),u.secret) {
        t.Fatal("unauthorized response must never disclose the secret")
    }

    bootstrap:=uiRequest(t,u,"GET","/?access="+u.secret,nil,false,"")
    if bootstrap.Code!=http.StatusSeeOther {
        t.Fatalf("private URL bootstrap: %d",bootstrap.Code)
    }
    cookieList:=bootstrap.Result().Cookies()
    if len(cookieList)!=1 {t.Fatalf("expected exactly one session cookie: %d",len(cookieList))}
    cookie:=cookieList[0]
    if cookie.Name!=u.sessionCookieName() || cookie.Value!=u.secret ||
       cookie.SameSite!=http.SameSiteLaxMode || !cookie.HttpOnly ||
       cookie.Path!="/" {
        t.Fatalf("unexpected bootstrap cookie attributes: %+v",cookie)
    }
    if bootstrap.Header().Get("Location")!="/" {t.Fatal("access token must be removed from the browser URL")}
    if bootstrap.Header().Get("Referrer-Policy")!="no-referrer" {
        t.Fatal("access-code redirect must not leak a referrer")
    }
    // Simulate Chrome following the safe top-level redirect with the new cookie.
    after:=httptest.NewRequest("GET","http://127.0.0.1:8765/",nil)
    after.AddCookie(cookie)
    response:=httptest.NewRecorder()
    u.ServeHTTP(response,after)
    if response.Code!=http.StatusOK || !strings.Contains(response.Body.String(),"Create an offline assessment") {
        t.Fatalf("session bootstrap did not reach assessment: %d",response.Code)
    }
    // A session cookie alone must not make a state-changing request safe.
    wrongOrigin:=uiRequest(t,u,"POST","/create",url.Values{
        "csrf":{u.secret},"scope":{"Synthetic"},"level":{"1"},
    },true,"https://untrusted.example")
    if wrongOrigin.Code!=http.StatusForbidden {t.Fatal("cross-origin POST accepted")}
    withoutFormToken:=uiRequest(t,u,"POST","/create",url.Values{
        "scope":{"Synthetic"},"level":{"1"},
    },true,"http://127.0.0.1:8765")
    if withoutFormToken.Code!=http.StatusConflict {t.Fatal("CSRF-less POST accepted")}
}

func TestUIFormDiagnosticsNoTokenLeakAndNoCrossOriginExecution(t *testing.T) {
    u:=testUI(t)
    cases:=[]struct{
        name string
        form url.Values
        origin string
        contentType string
        wantStatus int
        wantCode string
    }{
        {"stale_token",url.Values{"csrf":{"not-current"},"scope":{"Private Assessment"},"level":{"1"}},"http://127.0.0.1:8765","",http.StatusConflict,"FORM_TOKEN_EXPIRED"},
        {"missing_token",url.Values{"scope":{"Private Assessment"},"level":{"1"}},"","",http.StatusConflict,"FORM_TOKEN_EXPIRED"},
        {"cross_site",url.Values{"csrf":{u.secret},"scope":{"Private Assessment"},"level":{"1"}},"https://evil.example","",http.StatusForbidden,"ORIGIN_MISMATCH"},
        {"opaque_origin",url.Values{"csrf":{u.secret},"scope":{"Private Assessment"},"level":{"1"}},"null","",http.StatusForbidden,"ORIGIN_MISMATCH"},
        {"wrong_port",url.Values{"csrf":{u.secret},"scope":{"Private Assessment"},"level":{"1"}},"http://127.0.0.1:9999","",http.StatusForbidden,"ORIGIN_MISMATCH"},
        {"unsupported_type",url.Values{"csrf":{u.secret},"scope":{"Private Assessment"},"level":{"1"}},"","text/plain",http.StatusForbidden,"UNSUPPORTED_FORM_TYPE"},
    }
    for _,tc:=range cases {
        t.Run(tc.name,func(t *testing.T) {
            req:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8765/create",strings.NewReader(tc.form.Encode()))
            if tc.contentType!="" {req.Header.Set("Content-Type",tc.contentType)} else {req.Header.Set("Content-Type","application/x-www-form-urlencoded")}
            if tc.origin!="" {req.Header.Set("Origin",tc.origin)}
            req.AddCookie(&http.Cookie{Name:u.sessionCookieName(),Value:u.secret})
            out:=httptest.NewRecorder()
            u.ServeHTTP(out,req)
            if out.Code!=tc.wantStatus {t.Fatalf("got status %d, want %d: %s",out.Code,tc.wantStatus,out.Body.String())}
            if v:=out.Header().Get("X-OpenCMMC-Form-Diagnostic");v!=tc.wantCode {
                t.Fatalf("got diagnostic %q, want %q",v,tc.wantCode)
            }
            for _,needle:=range []string{tc.wantCode} {
                if !strings.Contains(out.Body.String(),needle) {t.Fatalf("missing diagnostic in response: %q",needle)}
            }
            for _,sensitive:=range []string{u.secret,"Private Assessment","https://evil.example"} {
                if strings.Contains(out.Body.String(),sensitive) {t.Fatal("unsafe form echo")}
            }
            if !strings.Contains(out.Header().Get("X-OpenCMMC-Version"),"dev") {
                t.Fatal("version must be identifiable even on failed requests")
            }
            if _,err:=os.Stat(filepath.Join(u.root,"state.json"));!os.IsNotExist(err) {
                t.Fatalf("rejected POST mutated local state: %v",err)
            }
        })
    }
    // A valid locally authenticated form must still work.
    valid:=uiRequest(t,u,"POST","/create",url.Values{
        "csrf":{u.secret},"scope":{"Synthetic"},"level":{"1"},
    },true,"http://127.0.0.1:8765")
    if valid.Code!=http.StatusSeeOther {
        t.Fatalf("valid form rejected: %d %s",valid.Code,valid.Body.String())
    }
    page:=uiRequest(t,u,"GET","/",nil,true,"")
    if page.Code!=http.StatusOK || !strings.Contains(page.Body.String(),"OpenCMMC build:") {
        t.Fatal("user-visible running version missing from workbench")
    }
}

func TestUIPortScopedSessionCookiesPreventParallelInstanceCollisions(t *testing.T) {
    // Chrome/ChromeOS cookies are not scoped to port. Two processes serving
    // 127.0.0.1:8765 and 127.0.0.1:8876 must have different cookie names.
    a:=testUI(t)
    b:=testUI(t)
    b.port="8876"
    b.secret=strings.Repeat("b",64)
    if a.sessionCookieName()==b.sessionCookieName() {
        t.Fatal("cookie names must be unique per TCP port")
    }
    if a.sessionCookieName()!="oc_session_p8765" {
        t.Fatalf("unexpected session-cookie name %q",a.sessionCookieName())
    }
    // A legacy unscoped cookie must not authenticate new builds.
    legacy:=httptest.NewRequest("GET","http://127.0.0.1:8765/",nil)
    legacy.AddCookie(&http.Cookie{Name:"oc_session",Value:a.secret})
    legacyResult:=httptest.NewRecorder()
    a.ServeHTTP(legacyResult,legacy)
    if legacyResult.Code!=http.StatusForbidden {
        t.Fatal("legacy cookie unexpectedly authenticated new build")
    }

    aAccess:=uiRequest(t,a,"GET","/?access="+a.secret,nil,false,"")
    if aAccess.Code!=http.StatusSeeOther{t.Fatalf("a bootstrap: %d",aAccess.Code)}
    // b's redirect, on its own port.
    bAccessReq:=httptest.NewRequest("GET","http://127.0.0.1:8876/?access="+b.secret,nil)
    bAccess:=httptest.NewRecorder()
    b.ServeHTTP(bAccess,bAccessReq)
    if bAccess.Code!=http.StatusSeeOther{t.Fatalf("b bootstrap: %d",bAccess.Code)}
    cookieA:=aAccess.Result().Cookies()[0]
    cookieB:=bAccess.Result().Cookies()[0]
    if cookieA.Name==cookieB.Name {t.Fatal("two instances issued colliding cookie names")}
    if cookieA.Value==cookieB.Value{t.Fatal("independent sessions reused a token")}
    if !cookieA.HttpOnly || !cookieB.HttpOnly ||
       cookieA.SameSite!=http.SameSiteLaxMode || cookieB.SameSite!=http.SameSiteLaxMode{
        t.Fatal("session cookie flags incorrect")
    }
    // A browser sends both host-scoped cookies, regardless of port.
    for _,tc:=range []struct{ui *localUI;port string}{
        {a,"8765"},{b,"8876"},
    } {
        req:=httptest.NewRequest("GET","http://127.0.0.1:"+tc.port+"/",nil)
        req.AddCookie(cookieA)
        req.AddCookie(cookieB)
        output:=httptest.NewRecorder()
        tc.ui.ServeHTTP(output,req)
        if output.Code!=http.StatusOK {
            t.Fatalf("instance on %s rejected the correct port-scoped cookie: %d",tc.port,output.Code)
        }
    }
    // A wrong-port cookie cannot authenticate a listener, even when it
    // carries a valid secret for a different process.
    req:=httptest.NewRequest("GET","http://127.0.0.1:8765/",nil)
    req.AddCookie(cookieB)
    output:=httptest.NewRecorder()
    a.ServeHTTP(output,req)
    if output.Code!=http.StatusForbidden {
        t.Fatal("accepted another port's session cookie")
    }
}

func TestUIOriginAllowedChromeOSAliasesAndRejectedOutsiders(t *testing.T) {
    // ChromeOS can forward a localhost or penguin.linux.test browser Origin
    // to the Crostini listener with Host rewritten to 127.0.0.1:8765.
    // Never accept a wildcard or a mismatched/non-HTTP origin.
    allowed:=[]string{
        "", "http://127.0.0.1:8765", "http://localhost:8765",
        "http://penguin.linux.test:8765", "http://[::1]:8765",
    }
    for _,origin:=range allowed {
        if !localUIOriginOK(origin,"127.0.0.1:8765","8765") {
            t.Errorf("failed to accept allowlisted browser Origin %q",origin)
        }
    }
    rejected:=[]string{
        "null", "file://", "chrome-extension://testing",
        "https://127.0.0.1:8765","http://127.0.0.1:9999",
        "http://localhost:8766","http://penguin.linux.test:9999",
        "http://evil.example:8765","http://evil.penguin.linux.test:8765",
        "http://localhost.evil.example:8765","http://0.0.0.0:8765",
        "http://192.168.0.8:8765","http://127.1:8765",
        "http://localhost:8765/","http://localhost:8765?secret=1",
        "http://localhost:8765#fragment","http://username@localhost:8765",
    }
    for _,origin:=range rejected {
        if localUIOriginOK(origin,"127.0.0.1:8765","8765") {
            t.Errorf("accepted untrusted or malformed Origin %q",origin)
        }
    }
    for _,host:=range []string{"evil.example:8765","localhost:8765","127.0.0.1:8766"} {
        if localUIOriginOK("http://localhost:8765",host,"8765") {
            t.Errorf("accepted untrusted Host %q",host)
        }
    }
}
func TestUIChromeOSForwardedOriginCanCreateAssessment(t *testing.T) {
    // Verify the real POST handler, not merely the Origin predicate.
    for _,origin:=range []string{
        "http://localhost:8765",
        "http://penguin.linux.test:8765",
        "http://[::1]:8765",
    } {
        t.Run(origin,func(t *testing.T){
            u:=testUI(t)
            post:=uiRequest(t,u,"POST","/create",url.Values{
                "scope":{"Synthetic ChromeOS loopback workspace"},
                "level":{"1"},
                "csrf":{u.secret},
            },true,origin)
            if post.Code!=http.StatusSeeOther {
                t.Fatalf("valid ChromeOS forwarded Origin rejected: %d %s",post.Code,post.Body.String())
            }
            if got:=post.Header().Get("X-OpenCMMC-Form-Diagnostic");got!="" {
                t.Fatalf("successful request has failure diagnostic: %q",got)
            }
            report:=uiRequest(t,u,"GET","/export",nil,true,"")
            if report.Code!=http.StatusOK {t.Fatalf("free local assessment export failed: %d",report.Code)}
            var output map[string]any
            if err:=json.Unmarshal(report.Body.Bytes(),&output);err!=nil {t.Fatal(err)}
        })
    }
}
func TestUIChromeOSAliasesStillRequireCSRFAndValidSession(t *testing.T) {
    for _,origin:=range []string{"http://localhost:8765","http://penguin.linux.test:8765"} {
        t.Run(origin,func(t *testing.T){
            u:=testUI(t)
            noCSRF:=uiRequest(t,u,"POST","/create",url.Values{
                "scope":{"Synthetic"},"level":{"1"},
            },true,origin)
            if noCSRF.Code!=http.StatusConflict ||
               noCSRF.Header().Get("X-OpenCMMC-Form-Diagnostic")!="FORM_TOKEN_EXPIRED" {
                t.Fatalf("CSRF-less request accepted: %d",noCSRF.Code)
            }
            noSession:=uiRequest(t,u,"POST","/create",url.Values{
                "scope":{"Synthetic"},"level":{"1"},"csrf":{u.secret},
            },false,origin)
            if noSession.Code!=http.StatusForbidden {t.Fatalf("unauthenticated POST accepted: %d",noSession.Code)}
            if _,err:=os.Stat(filepath.Join(u.root,"state.json"));!os.IsNotExist(err) {
                t.Fatalf("rejected request wrote assessment: %v",err)
            }
        })
    }
}

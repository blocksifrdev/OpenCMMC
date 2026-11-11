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
    if authenticated {req.AddCookie(&http.Cookie{Name:"oc_session",Value:u.secret})}
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
    badHost.AddCookie(&http.Cookie{Name:"oc_session",Value:u.secret})
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
    req.AddCookie(&http.Cookie{Name:"oc_session",Value:u.secret})
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
    if cookie.Name!="oc_session" || cookie.Value!=u.secret ||
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

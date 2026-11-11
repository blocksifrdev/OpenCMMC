package main

import (
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
    "testing"
)

func createHandoffTestWorkspace(t *testing.T,u *localUI,scope string) {
    t.Helper()
    r:=uiRequest(t,u,"POST","/create",url.Values{
        "scope":{scope},"level":{"1"},"csrf":{u.secret},
    },true,"http://127.0.0.1:8765")
    if r.Code!=http.StatusSeeOther {t.Fatalf("create workspace: %d %s",r.Code,r.Body.String())}
}
func TestHandoffRequiresLocalWorkspace(t *testing.T) {
    u:=testUI(t)
    if got:=uiRequest(t,u,"GET","/handoff?service=blake",nil,false,"").Code;got!=http.StatusForbidden {t.Fatalf("unauthenticated %d",got)}
    if got:=uiRequest(t,u,"GET","/handoff?service=blake",nil,true,"").Code;got!=http.StatusBadRequest {t.Fatalf("no workspace %d",got)}
    if got:=uiRequest(t,u,"POST","/handoff",url.Values{
        "csrf":{u.secret},"service":{"blake"},"confirm":{"yes"},
    },true,"").Code;got!=http.StatusBadRequest {t.Fatalf("pre-workspace export %d",got)}
}
func TestHandoffServiceConsentAndPrivacy(t *testing.T) {
    for _,tc:=range []struct{ID,Name,Workflow string}{
        {"cortextrace","CortexTrace","GOVERN_AI_CODING_SESSION"},
        {"control","BlockSiFr Control","EVALUATE_RUNTIME_AUTHORITY"},
        {"blake","BLAKE","REQUEST_AUTHORIZED_VULNERABILITY_DISCOVERY"},
    } {
        t.Run(tc.ID,func(t *testing.T){
            u:=testUI(t)
            scope:="SENSITIVE CUSTOMER SYSTEM DO NOT EXPORT"
            createHandoffTestWorkspace(t,u,scope)
            page:=uiRequest(t,u,"GET","/handoff?service="+tc.ID,nil,true,"")
            if page.Code!=http.StatusOK {t.Fatalf("local page %d: %s",page.Code,page.Body.String())}
            body:=page.Body.String()
            for _,v:=range []string{tc.Name,"Download private request","No data is submitted","noopener noreferrer","Return to free OpenCMMC assessment"} {
                if !strings.Contains(body,v) {t.Errorf("missing %q",v)}
            }
            if strings.Contains(body,scope) || strings.Contains(body,"<script") || strings.Contains(body,"<iframe") {
                t.Fatal("handoff page leaked scope or loaded active remote content")
            }
            if page.Header().Get("Referrer-Policy")!="no-referrer" {t.Fatal("missing no-referrer")}
            missing:=uiRequest(t,u,"POST","/handoff",url.Values{"csrf":{u.secret},"service":{tc.ID}},true,"")
            if missing.Code!=http.StatusBadRequest {t.Fatalf("unconfirmed request generated: %d",missing.Code)}
            confirmed:=uiRequest(t,u,"POST","/handoff",url.Values{
                "csrf":{u.secret},"service":{tc.ID},"confirm":{"yes"},
            },true,"http://127.0.0.1:8765")
            if confirmed.Code!=http.StatusOK {t.Fatalf("confirm %d: %s",confirmed.Code,confirmed.Body.String())}
            if !strings.Contains(confirmed.Header().Get("Content-Disposition"),"opencmmc-"+tc.ID+"-request.json") {
                t.Fatal("request must download as local JSON")
            }
            if strings.Contains(confirmed.Body.String(),scope) || strings.Contains(confirmed.Body.String(),u.secret) {
                t.Fatal("local request exported assessment scope or token")
            }
            var m map[string]any
            if err:=json.Unmarshal(confirmed.Body.Bytes(),&m);err!=nil {t.Fatal(err)}
            expectations:=map[string]any{
                "schema_version":"opencmmc.service-intent.v1","service":tc.ID,
                "requested_workflow":tc.Workflow,"status":"PREPARED_LOCAL_ONLY",
                "network_transfer_performed":false,"service_activated":false,
                "assessment_data_included":false,"repository_access_granted":false,
                "customer_authorization":"REQUIRED_SEPARATELY","local_intent_confirmation":true,
            }
            for k,want:=range expectations {if m[k]!=want {t.Errorf("%s got %v wanted %v",k,m[k],want)}}
            if len(m["requested_evidence"].([]any))!=0 || len(m["selected_assets"].([]any))!=0 {
                t.Fatal("handoff included targets or evidence")
            }
            if id,ok:=m["request_id"].(string);!ok || len(id)!=36 || !strings.HasPrefix(id,"osh_") {
                t.Fatalf("invalid random reference: %v",m["request_id"])
            }
            free:=uiRequest(t,u,"GET","/export",nil,true,"")
            if free.Code!=http.StatusOK {t.Fatalf("free export broken: %d",free.Code)}
            if strings.Contains(free.Body.String(),"requested_workflow") {
                t.Fatal("commercial handoff modified the free readiness report")
            }
        })
    }
}
func TestHandoffRejectsInvalidInputs(t *testing.T) {
    u:=testUI(t)
    createHandoffTestWorkspace(t,u,"Synthetic boundary")
    for _,path:=range []string{
        "/handoff","/handoff?service=unknown","/handoff?service=control&scope=secret",
        "/handoff?service=cortextrace&service=blake",
    } {
        if code:=uiRequest(t,u,"GET",path,nil,true,"").Code;code!=http.StatusBadRequest {
            t.Errorf("invalid GET %s accepted: %d",path,code)
        }
    }
    for _,tc:=range []struct{Name,Path,Origin string; Form url.Values;Code int}{
        {"bad_csrf","/handoff","",url.Values{"csrf":{"bad"},"service":{"blake"},"confirm":{"yes"}},http.StatusConflict},
        {"cross_origin","/handoff","https://evil.example",url.Values{"csrf":{u.secret},"service":{"blake"},"confirm":{"yes"}},http.StatusForbidden},
        {"unknown_service","/handoff","",url.Values{"csrf":{u.secret},"service":{"unknown"},"confirm":{"yes"}},http.StatusBadRequest},
        {"unexpected_field","/handoff","",url.Values{"csrf":{u.secret},"service":{"blake"},"confirm":{"yes"},"scope":{"secret"}},http.StatusBadRequest},
        {"duplicate_service","/handoff","",url.Values{"csrf":{u.secret},"service":{"blake","control"},"confirm":{"yes"}},http.StatusBadRequest},
        {"query_parameter","/handoff?scope=secret","",url.Values{"csrf":{u.secret},"service":{"blake"},"confirm":{"yes"}},http.StatusBadRequest},
    } {
        t.Run(tc.Name,func(t *testing.T){
            r:=uiRequest(t,u,"POST",tc.Path,tc.Form,true,tc.Origin)
            if r.Code!=tc.Code {t.Fatalf("got %d, want %d: %s",r.Code,tc.Code,r.Body.String())}
            if strings.Contains(r.Header().Get("Content-Disposition"),".json") {t.Fatal("rejected request must not create a file")}
        })
    }
}

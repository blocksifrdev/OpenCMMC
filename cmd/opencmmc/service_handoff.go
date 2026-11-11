package main

// Offline handoff preparation. These routes do not call any remote service,
// open sockets, transmit evidence, or activate a commercial integration.
import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "html/template"
    "net/http"
    "strings"
    "time"
)

type serviceHandoff struct {
    ID string
    Name string
    Question string
    Purpose string
    Workflow string
    URL string
    NextStep string
}

type handoffPage struct {
    CSRF string
    Service serviceHandoff
}

func serviceForHandoff(id string) (serviceHandoff, bool) {
    switch id {
    case "cortextrace":
        return serviceHandoff{
            ID: "cortextrace",
            Name: "CortexTrace",
            Question: "Would you like to govern remediation in an AI coding session?",
            Purpose: "Change impact, approvals and execution evidence for AI-assisted software delivery.",
            Workflow: "GOVERN_AI_CODING_SESSION",
            URL: "https://exchange.blocksifr.com/solutions/software-delivery-governance",
            NextStep: "Review the service with your team; a separate authenticated session and repository authorization are required.",
        }, true
    case "control":
        return serviceHandoff{
            ID: "control",
            Name: "BlockSiFr Control",
            Question: "Would you like to control automated changes and agent actions?",
            Purpose: "Establish consequence-aware runtime authority, step-up decisions and execution receipts.",
            Workflow: "EVALUATE_RUNTIME_AUTHORITY",
            URL: "https://www.blocksifr.com/control",
            NextStep: "Define a protected action and arrange separate tenant onboarding and authority policy approval.",
        }, true
    case "blake":
        return serviceHandoff{
            ID: "blake",
            Name: "BLAKE",
            Question: "Would you like deeper, consequence-focused vulnerability discovery?",
            Purpose: "Request a scoped security investigation beyond CMMC readiness evidence review.",
            Workflow: "REQUEST_AUTHORIZED_VULNERABILITY_DISCOVERY",
            URL: "https://www.blocksifr.com/",
            NextStep: "Establish written asset ownership, scope and testing permission with the service before any assessment starts.",
        }, true
    default:
        return serviceHandoff{}, false
    }
}

func (u *localUI) checkHandoffWorkspace() error {
    state,err:=u.state()
    if err!=nil {return err}
    if state.Mode!="basic" && state.Mode!="objectives" {
        return errors.New("create a local assessment before preparing an optional service request")
    }
    return nil
}

func (u *localUI) viewHandoff(w http.ResponseWriter,r *http.Request) error {
    if err:=u.checkHandoffWorkspace();err!=nil{return err}
    if len(r.URL.Query())!=1 || len(r.URL.Query()["service"])!=1 {
        return errors.New("choose exactly one optional service")
    }
    svc,ok:=serviceForHandoff(r.URL.Query().Get("service"))
    if !ok{return errors.New("unknown optional service")}
    w.Header().Set("Content-Type","text/html; charset=utf-8")
    return handoffTemplate.Execute(w,handoffPage{CSRF:u.secret,Service:svc})
}

func (u *localUI) prepareHandoff(w http.ResponseWriter,r *http.Request) error {
    if err:=u.checkHandoffWorkspace();err!=nil{return err}
    if r.URL.RawQuery!=""{return errors.New("unexpected request parameters")}
    if err:=r.ParseForm();err!=nil{return errors.New("invalid handoff form")}
    // A handoff is intentionally built from allowlisted fields only.
    for key := range r.PostForm {
        switch key {
        case "csrf","service","confirm":
        default: return errors.New("unexpected form field")
        }
        if len(r.PostForm[key])!=1 {return errors.New("duplicate form field")}
    }
    if r.PostFormValue("confirm")!="yes" {
        return errors.New("confirm local-only preparation before downloading a request")
    }
    svc,ok:=serviceForHandoff(r.PostFormValue("service"))
    if !ok {return errors.New("unknown optional service")}
    requestID:=make([]byte,16)
    if _,err:=rand.Read(requestID);err!=nil{return err}
    // Not derived from the assessment: no CUI, scope, control ID, finding,
    // filename, hash, reviewer, credentials, or private session URL.
    payload:=map[string]any{
        "schema_version":"opencmmc.service-intent.v1",
        "request_id":"osh_"+hex.EncodeToString(requestID),
        "created_at":time.Now().UTC().Format(time.RFC3339),
        "source":"OpenCMMC",
        "service":svc.ID,
        "requested_workflow":svc.Workflow,
        "status":"PREPARED_LOCAL_ONLY",
        "service_activated":false,
        "network_transfer_performed":false,
        "assessment_data_included":false,
        "repository_access_granted":false,
        "customer_authorization":"REQUIRED_SEPARATELY",
        "local_intent_confirmation":true,
        "requested_evidence":[]string{},
        "selected_assets":[]string{},
        "provider_information_url":svc.URL,
        "next_step":svc.NextStep,
        "notice":"This file is an offline planning request, not an API submission, purchased service, authorized scan or enforcement action.",
    }
    out,err:=json.MarshalIndent(payload,"","  ")
    if err!=nil{return err}
    w.Header().Set("Content-Type","application/json; charset=utf-8")
    w.Header().Set("Content-Disposition",fmt.Sprintf("attachment; filename=%q","opencmmc-"+svc.ID+"-request.json"))
    _,err=w.Write(append(out,'\n'))
    return err
}

var handoffTemplate = template.Must(template.New("optional-service-handoff").Parse(
"<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\">\n" +
"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n" +
"<title>{{.Service.Name}} | OpenCMMC local request</title>\n" +
"<style>\n" +
":root{font:16px system-ui,sans-serif;background:#f7f5f0;color:#1c1c1c}\n" +
"body{max-width:720px;margin:38px auto;padding:0 18px;line-height:1.55}\n" +
"main{background:#fff;border:1px solid #dad7d0;border-radius:12px;padding:26px}\n" +
"h1{font-size:25px}h2{font-size:19px}\n" +
".notice{padding:14px;border:1px solid #ded4bf;background:#faf7ef;border-radius:7px}\n" +
"label{display:block;margin:18px 0}input{margin-right:8px}\n" +
"button{background:#181818;color:#fff;border:0;border-radius:6px;padding:12px 16px;font:inherit;cursor:pointer}\n" +
"a{color:#1c4c75}\n" +
"small{color:#565656}\n" +
"</style></head><body><main>\n" +
"<p><a href=\"/\">Return to free OpenCMMC assessment</a></p>\n" +
"<h1>{{.Service.Name}}: private next-step request</h1>\n" +
"<p><strong>{{.Service.Question}}</strong></p>\n" +
"<p>{{.Service.Purpose}}</p>\n" +
"<div class=\"notice\"><strong>Nothing will be shared.</strong> You can prepare an optional local JSON request without including your assessment, findings, evidence, system scope or credentials. This action does not contact BlockSiFr, activate a service, enroll an account, or authorize testing. OpenCMMC remains free whether or not you proceed.</div>\n" +
"<h2>What the request contains</h2>\n" +
"<p>Only the selected service name, a predefined purpose, a random request reference, the preparation time and clear statements that no service has been activated. No assessment or customer data.</p>\n" +
"<form method=\"post\" action=\"/handoff\">\n" +
"<input type=\"hidden\" name=\"csrf\" value=\"{{.CSRF}}\">\n" +
"<input type=\"hidden\" name=\"service\" value=\"{{.Service.ID}}\">\n" +
"<label><input type=\"checkbox\" name=\"confirm\" value=\"yes\" required>I understand this only downloads a local planning file. No data is submitted and no paid service starts.</label>\n" +
"<button type=\"submit\">Download private request</button></form>\n" +
"<h2>Next step</h2>\n" +
"<p>{{.Service.NextStep}}</p>\n" +
"<p><a href=\"{{.Service.URL}}\" target=\"_blank\" rel=\"noopener noreferrer\">Open the {{.Service.Name}} information page</a></p>\n" +
"<p><small>Opening the external page is a separate, deliberate action requiring network access. Obtain organizational authorization before sharing any sensitive information with a provider. No workspace details are included in this link.</small></p>\n" +
"</main></body></html>\n"))

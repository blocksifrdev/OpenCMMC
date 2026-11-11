package main

import (
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
    "testing"

    "github.com/blocksifrdev/OpenCMMC/internal/cmmc"
)

func TestOptionalServicesOnlyAfterWorkspaceAndNoAutomaticTransfer(t *testing.T) {
    u:=testUI(t)
    before:=uiRequest(t,u,"GET","/",nil,true,"")
    if before.Code!=http.StatusOK { t.Fatalf("start page status: %d",before.Code) }
    if strings.Contains(before.Body.String(),"optional-services") { t.Fatal("outbound offers must wait until user creates a workspace") }

    create:=uiRequest(t,u,"POST","/create",url.Values{
        "scope":{"Synthetic project, not real evidence"},
        "level":{"1"},
        "csrf":{u.secret},
    },true,"http://127.0.0.1:8765")
    if create.Code!=http.StatusSeeOther { t.Fatalf("create: %d",create.Code) }

    result:=uiRequest(t,u,"GET","/",nil,true,"")
    if result.Code!=http.StatusOK { t.Fatalf("workbench: %d %s", result.Code,result.Body.String()) }
    page:=result.Body.String()
    for _, required:=range []string{
        "Want to go further?",
        "CortexTrace",
        "BlockSiFr Control",
        "BLAKE",
        "https://exchange.blocksifr.com/solutions/software-delivery-governance",
        "https://www.blocksifr.com/control",
        "https://www.blocksifr.com/",
        `target="_blank" rel="noopener noreferrer"`,
        "No workspace, control identifier, evidence, file path or report is sent",
    } {
        if !strings.Contains(page,required) { t.Errorf("missing optional link or disclosure: %q",required) }
    }
    if strings.Contains(page,"Synthetic project, not real evidence") && strings.Contains(page,"?scope=") {
        t.Fatal("outbound query string must not include workspace scope")
    }
    if strings.Contains(page,`<script`) || strings.Contains(page,`<iframe`) {
        t.Fatal("optional services must not install remote scripts or frames")
    }
    if got:=result.Header().Get("Referrer-Policy");got!="no-referrer" { t.Fatalf("referrer policy: %q",got) }
    if !strings.Contains(result.Header().Get("Content-Security-Policy"),"default-src 'none'") {
        t.Fatal("external scripts/requests must stay blocked")
    }

    exp:=uiRequest(t,u,"GET","/export",nil,true,"")
    if exp.Code!=http.StatusOK {t.Fatalf("export: %d",exp.Code)}
    var report cmmc.Report
    if err:=json.Unmarshal(exp.Body.Bytes(),&report);err!=nil {t.Fatal(err)}
    if report.Summary.Unassessed!=15 { t.Fatalf("optional service offer changed free readiness: %+v",report.Summary) }
    if strings.Contains(exp.Body.String(),"blocksifr.com") { t.Fatal("export must contain only assessment evidence, not promotion") }
}

func TestOptionalServiceLinksHaveNoSensitiveQueryParameters(t *testing.T) {
    u:=testUI(t)
    create:=uiRequest(t,u,"POST","/create",url.Values{
        "scope":{"Protected local SECRET SCOPE"},
        "level":{"1"},
        "csrf":{u.secret},
    },true,"")
    if create.Code!=http.StatusSeeOther { t.Fatalf("create: %d",create.Code)}
    page:=uiRequest(t,u,"GET","/",nil,true,"").Body.String()
    for _, link:=range []string{
        "https://exchange.blocksifr.com/solutions/software-delivery-governance",
        "https://www.blocksifr.com/control",
        "https://www.blocksifr.com/",
    }{
        if !strings.Contains(page,`href="`+link+`"`) {t.Errorf("exact outbound URL missing: %s",link)}
    }
    if strings.Contains(page,"?scope=") || strings.Contains(page,"?control=") && strings.Contains(page,"blocksifr.com?") {
        t.Fatal("unexpected sensitive outbound query parameters")
    }
}

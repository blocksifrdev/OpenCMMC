package main

import (
    "bytes"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "testing"

    "github.com/blocksifrdev/OpenCMMC/internal/cmmc"
)

// If OpenCMMC is interrupted after writing the workspace but before state.json,
// reopening must recover the existing assessment without deleting or changing it.
func TestUIRecoversInterruptedCreationWithoutOverwriting(t *testing.T) {
    for _,mode:=range []string{"basic","objectives"} {
        t.Run(mode,func(t *testing.T){
            u:=testUI(t)
            var path string
            if mode=="basic" {
                ws,err:=cmmc.NewWorkspace(1,"Synthetic interrupted Level 1 workspace")
                if err!=nil {t.Fatal(err)}
                path=filepath.Join(u.root,"requirements.json")
                if err=uiWriteJSON(path,ws);err!=nil {t.Fatal(err)}
            } else {
                ws,err:=cmmc.NewObjectiveWorkspace(2,"Synthetic interrupted Level 2 workspace")
                if err!=nil {t.Fatal(err)}
                path=filepath.Join(u.root,"objectives.json")
                if err=uiWriteJSON(path,ws);err!=nil {t.Fatal(err)}
            }
            original,err:=os.ReadFile(path)
            if err!=nil {t.Fatal(err)}
            if _,err=os.Stat(filepath.Join(u.root,"state.json"));!os.IsNotExist(err) {
                t.Fatalf("fixture unexpectedly has state.json: %v",err)
            }

            recovered,err:=u.state()
            if err!=nil {t.Fatalf("valid orphan was not recovered: %v",err)}
            if recovered.Mode!=mode {t.Fatalf("recovered mode %q, expected %q",recovered.Mode,mode)}
            if recovered.Level!=map[string]int{"basic":1,"objectives":2}[mode] {
                t.Fatalf("unexpected recovered level: %d",recovered.Level)
            }
            view:=uiRequest(t,u,http.MethodGet,"/",nil,true,"")
            if view.Code!=http.StatusOK {
                t.Fatalf("recovered workspace not viewable: %d %s",view.Code,view.Body.String())
            }
            export:=uiRequest(t,u,http.MethodGet,"/export",nil,true,"")
            if export.Code!=http.StatusOK {
                t.Fatalf("recovered workspace not exportable: %d %s",export.Code,export.Body.String())
            }
            overwrite:=uiRequest(t,u,http.MethodPost,"/create",url.Values{
                "scope":{"An attempted replacement"},
                "level":{"1"},
                "csrf":{u.secret},
            },true,"http://127.0.0.1:8765")
            if overwrite.Code!=http.StatusBadRequest {
                t.Fatalf("existing assessment was not protected against replacement: %d",overwrite.Code)
            }
            after,err:=os.ReadFile(path)
            if err!=nil {t.Fatal(err)}
            if !bytes.Equal(original,after) {t.Fatal("existing assessment changed during recovery or rejected replacement")}
            if _,err=os.Stat(filepath.Join(u.root,"state.json"));!os.IsNotExist(err) {
                t.Fatalf("read-only recovery should not write new state: %v",err)
            }
        })
    }
}

func TestUIFailsClosedForCorruptOrAmbiguousInterruptedAssessment(t *testing.T) {
    t.Run("corrupt",func(t *testing.T){
        u:=testUI(t)
        path:=filepath.Join(u.root,"requirements.json")
        original:=[]byte("{invalid assessment")
        if err:=os.WriteFile(path,original,0600);err!=nil {t.Fatal(err)}
        if _,err:=u.state();err==nil {t.Fatal("corrupt orphan was accepted")}
        attempted:=uiRequest(t,u,http.MethodPost,"/create",url.Values{
            "scope":{"Replacement"},"level":{"1"},"csrf":{u.secret},
        },true,"http://127.0.0.1:8765")
        if attempted.Code!=http.StatusBadRequest {t.Fatalf("corrupt assessment overwritten: %d",attempted.Code)}
        current,err:=os.ReadFile(path)
        if err!=nil {t.Fatal(err)}
        if !bytes.Equal(original,current) {t.Fatal("corrupt original was changed")}
    })
    t.Run("ambiguous",func(t *testing.T){
        u:=testUI(t)
        basic,err:=cmmc.NewWorkspace(1,"Synthetic Level 1")
        if err!=nil {t.Fatal(err)}
        objective,err:=cmmc.NewObjectiveWorkspace(2,"Synthetic Level 2")
        if err!=nil {t.Fatal(err)}
        if err=uiWriteJSON(filepath.Join(u.root,"requirements.json"),basic);err!=nil {t.Fatal(err)}
        if err=uiWriteJSON(filepath.Join(u.root,"objectives.json"),objective);err!=nil {t.Fatal(err)}
        if _,err=u.state();err==nil {t.Fatal("ambiguous orphan was auto-selected")}
        attempted:=uiRequest(t,u,http.MethodPost,"/create",url.Values{
            "scope":{"Replacement"},"level":{"2"},"csrf":{u.secret},
        },true,"http://127.0.0.1:8765")
        if attempted.Code!=http.StatusBadRequest {t.Fatalf("ambiguous assessment overwritten: %d",attempted.Code)}
    })
}

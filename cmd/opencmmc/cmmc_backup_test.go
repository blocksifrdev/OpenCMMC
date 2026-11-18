package main

import (
    "archive/zip"
    "bytes"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/blocksifrdev/OpenCMMC/internal/cmmc"
)

func syntheticBackupWorkspace(t *testing.T,level int,withEvidence bool) string {
    t.Helper()
    root,err:=initUIRoot(t.TempDir())
    if err!=nil{t.Fatal(err)}
    if level==1 {
        w,err:=cmmc.NewWorkspace(1,"Synthetic offline backup workspace")
        if err!=nil{t.Fatal(err)}
        if withEvidence {
            b:=[]byte("Synthetic evidence only. No CUI or customer data.\n")
            digest:=sha256.Sum256(b)
            name:="evidence-test.bin"
            if err=os.WriteFile(filepath.Join(root,"evidence",name),b,0600);err!=nil{t.Fatal(err)}
            w.Controls[0].Evidence=append(w.Controls[0].Evidence,cmmc.Evidence{
                Path:name,SHA256:hex.EncodeToString(digest[:]),CapturedAt:time.Now().UTC().Format(time.RFC3339),
            })
        }
        if err=cmmc.ValidatedWorkspace(w);err!=nil{t.Fatal(err)}
        if err=uiWriteJSON(filepath.Join(root,"requirements.json"),w);err!=nil{t.Fatal(err)}
        if err=uiWriteJSON(filepath.Join(root,"state.json"),localUIState{Mode:"basic",Level:1});err!=nil{t.Fatal(err)}
    } else {
        w,err:=cmmc.NewObjectiveWorkspace(2,"Synthetic offline backup Level 2 workspace")
        if err!=nil{t.Fatal(err)}
        if err=uiWriteJSON(filepath.Join(root,"objectives.json"),w);err!=nil{t.Fatal(err)}
        if err=uiWriteJSON(filepath.Join(root,"state.json"),localUIState{Mode:"objectives",Level:2});err!=nil{t.Fatal(err)}
    }
    return root
}

func TestOfflineBackupAndRestoreRoundTrip(t *testing.T) {
    for _,tc:=range []struct{level int;withEvidence bool}{
        {1,true},{2,false},
    }{
        t.Run(map[int]string{1:"level1",2:"level2"}[tc.level],func(t *testing.T){
            root:=syntheticBackupWorkspace(t,tc.level,tc.withEvidence)
            archive:=filepath.Join(t.TempDir(),"assessment.zip")
            if err:=snapshotCreate(root,archive);err!=nil{t.Fatal(err)}
            before,err:=snapshotLoad(archive)
            if err!=nil{t.Fatal(err)}
            dest:=filepath.Join(t.TempDir(),"restored")
            if err=snapshotRestore(archive,dest);err!=nil{t.Fatal(err)}
            after,err:=snapshotCollect(dest)
            if err!=nil{t.Fatal(err)}
            if len(after)!=len(before){t.Fatalf("snapshot files differ: %d, %d",len(before),len(after))}
            for name,initial:=range before {
                if !bytes.Equal(initial,after[name]){t.Fatalf("file changed on restore: %s",name)}
            }
            // Never overwrite original or already restored data.
            if err=snapshotCreate(root,archive);err==nil{t.Fatal("existing backup overwritten")}
            if err=snapshotRestore(archive,dest);err==nil{t.Fatal("existing assessment overwritten")}
            if _,err=os.Stat(filepath.Join(dest,"state.json"));err!=nil{t.Fatal("restore unexpectedly deleted")}
        })
    }
}
func TestOfflineBackupRejectsAlteredEvidenceAndSymlink(t *testing.T){
    root:=syntheticBackupWorkspace(t,1,true)
    if err:=os.WriteFile(filepath.Join(root,"evidence","evidence-test.bin"),[]byte("tampered"),0600);err!=nil{t.Fatal(err)}
    archive:=filepath.Join(t.TempDir(),"bad.zip")
    if err:=snapshotCreate(root,archive);err==nil{t.Fatal("tampered evidence was archived")}
    if _,err:=os.Stat(archive);!os.IsNotExist(err){t.Fatalf("invalid archive was left behind: %v",err)}

    root=syntheticBackupWorkspace(t,1,true)
    path:=filepath.Join(root,"evidence","evidence-test.bin")
    if err:=os.Remove(path);err!=nil{t.Fatal(err)}
    target:=filepath.Join(t.TempDir(),"not-evidence")
    if err:=os.WriteFile(target,[]byte("Synthetic evidence only. No CUI or customer data.\n"),0600);err!=nil{t.Fatal(err)}
    if err:=os.Symlink(target,path);err!=nil{t.Skipf("symlink creation unsupported: %v",err)}
    if err:=snapshotCreate(root,filepath.Join(t.TempDir(),"symlink.zip"));err==nil{
        t.Fatal("symlinked evidence was accepted")
    }
}
func TestOfflineBackupManifestTamperingFailsBeforeRestore(t *testing.T) {
    root:=syntheticBackupWorkspace(t,1,true)
    files,err:=snapshotCollect(root)
    if err!=nil{t.Fatal(err)}
    manifest:=snapshotBuildManifest(files)
    payload,err:=json.Marshal(manifest)
    if err!=nil{t.Fatal(err)}
    archive:=filepath.Join(t.TempDir(),"altered.zip")
    output,err:=os.Create(archive)
    if err!=nil{t.Fatal(err)}
    writer:=zip.NewWriter(output)
    for _,record:=range append([]snapshotEntry{{Name:"MANIFEST.json"}},manifest.Files...) {
        header:=&zip.FileHeader{Name:record.Name,Method:zip.Store}
        header.SetMode(0600)
        entry,e:=writer.CreateHeader(header)
        if e!=nil{t.Fatal(e)}
        b:=files[record.Name]
        if record.Name=="MANIFEST.json"{b=payload}
        if record.Name=="state.json"{b=append([]byte(nil),b...);b[0]='X'}
        if _,e=entry.Write(b);e!=nil{t.Fatal(e)}
    }
    if err=writer.Close();err!=nil{t.Fatal(err)}
    if err=output.Close();err!=nil{t.Fatal(err)}
    if _,err=snapshotLoad(archive);err==nil{t.Fatal("tampered backup accepted")}
    dest:=filepath.Join(t.TempDir(),"restore")
    if err=snapshotRestore(archive,dest);err==nil{t.Fatal("tampered backup restored")}
    if _,err=os.Stat(dest);!os.IsNotExist(err){t.Fatal("restore destination created from tampered archive")}
}
func TestOfflineBackupRejectsZipSlipAndAmbiguousWorkspace(t *testing.T){
    archive:=filepath.Join(t.TempDir(),"path.zip")
    out,err:=os.Create(archive)
    if err!=nil{t.Fatal(err)}
    writer:=zip.NewWriter(out)
    head:=&zip.FileHeader{Name:"../outside.txt",Method:zip.Store}
    head.SetMode(0600)
    entry,err:=writer.CreateHeader(head)
    if err!=nil{t.Fatal(err)}
    if _,err=entry.Write([]byte("not allowed"));err!=nil{t.Fatal(err)}
    if err=writer.Close();err!=nil{t.Fatal(err)}
    if err=out.Close();err!=nil{t.Fatal(err)}
    if _,err=snapshotLoad(archive);err==nil{t.Fatal("zip-slip payload accepted")}

    root:=syntheticBackupWorkspace(t,1,false)
    other,err:=cmmc.NewObjectiveWorkspace(2,"Synthetic duplicate")
    if err!=nil{t.Fatal(err)}
    if err=uiWriteJSON(filepath.Join(root,"objectives.json"),other);err!=nil{t.Fatal(err)}
    if err=snapshotCreate(root,filepath.Join(t.TempDir(),"ambiguous.zip"));err==nil{
        t.Fatal("ambiguous assessment was archived")
    }
}
func TestOfflineBackupCLIOutputAvoidsSecrets(t *testing.T){
    root:=syntheticBackupWorkspace(t,1,false)
    archive:=filepath.Join(t.TempDir(),"safe.zip")
    var out bytes.Buffer
    if err:=cmmcBackupCommand([]string{"backup","--data-dir",root,"--archive",archive},&out);err!=nil{t.Fatal(err)}
    if !bytes.Contains(out.Bytes(),[]byte("NOT ENCRYPTED")){t.Fatal("plaintext archive warning absent")}
    out.Reset()
    if err:=cmmcBackupCommand([]string{"verify-backup","--archive",archive},&out);err!=nil{t.Fatal(err)}
    out.Reset()
    if err:=cmmcBackupCommand([]string{"restore-backup","--archive",archive,"--out-dir",filepath.Join(t.TempDir(),"restore")},&out);err!=nil{t.Fatal(err)}
}

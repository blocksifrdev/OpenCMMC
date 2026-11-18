package main

// Offline OpenCMMC assessment snapshots. These archives are plaintext:
// operators MUST store them on approved encrypted media and keep the app
// stopped while backing up. Verification proves byte integrity only.
import (
    "archive/zip"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/blocksifrdev/OpenCMMC/internal/cmmc"
)

const snapshotSchema = "opencmmc.backup.v1"
const snapshotMaxFile int64 = 50 << 20
const snapshotMaxTotal int64 = 512 << 20
const snapshotMaxFiles = 4096

type snapshotEntry struct {
    Name   string
    SHA256 string
    Size   int64
}
type snapshotManifest struct {
    Schema string
    Files []snapshotEntry
}
func snapshotDigest(b []byte) string {
    sum:=sha256.Sum256(b)
    return hex.EncodeToString(sum[:])
}
func snapshotIsFileName(name string) bool {
    if name=="state.json" || name=="requirements.json" || name=="objectives.json" {return true}
    if strings.HasPrefix(name,"evidence/") && strings.HasSuffix(name,".bin") {
        base:=strings.TrimPrefix(name,"evidence/")
        return len(base)>len("evidence-.bin") && strings.HasPrefix(base,"evidence-") &&
            filepath.Base(base)==base && !strings.ContainsAny(base,"/\\")
    }
    return false
}
func snapshotReadFile(root,name string,limit int64) ([]byte,error) {
    if !snapshotIsFileName(name) {return nil,fmt.Errorf("unsafe snapshot entry %q",name)}
    // Reject symlinked directories and files, nonregular files and oversized
    // data. Recheck identity after opening to catch ordinary swap races.
    parts:=strings.Split(name,"/")
    dir:=root
    for _,part:=range parts[:len(parts)-1] {
        dir=filepath.Join(dir,part)
        st,err:=os.Lstat(dir)
        if err!=nil{return nil,err}
        if !st.IsDir() || st.Mode()&os.ModeSymlink!=0 {return nil,errors.New("unsafe evidence directory")}
    }
    path:=filepath.Join(root,filepath.FromSlash(name))
    st,err:=os.Lstat(path)
    if err!=nil{return nil,err}
    if !st.Mode().IsRegular() || st.Mode()&os.ModeSymlink!=0 || st.Size()>limit {
        return nil,fmt.Errorf("snapshot file is unsafe or too large: %s",name)
    }
    f,err:=os.Open(path)
    if err!=nil{return nil,err}
    defer f.Close()
    opened,err:=f.Stat()
    if err!=nil{return nil,err}
    if !opened.Mode().IsRegular() || !os.SameFile(st,opened) {return nil,errors.New("snapshot file changed during open")}
    b,err:=io.ReadAll(io.LimitReader(f,limit+1))
    if err!=nil{return nil,err}
    if int64(len(b))>limit || int64(len(b))!=opened.Size() {
        return nil,errors.New("snapshot file changed or exceeded limit during read")
    }
    return b,nil
}
func snapshotRequiredEvidence(files map[string][]byte) (map[string]string,error) {
    stateBytes,ok:=files["state.json"]
    if !ok{return nil,errors.New("snapshot missing state.json; do not guess workspace mode")}
    var state localUIState
    if err:=json.Unmarshal(stateBytes,&state);err!=nil{return nil,err}
    wanted:=map[string]string{}
    if state.Mode=="basic" && state.Level==1 {
        data,ok:=files["requirements.json"]
        if !ok || files["objectives.json"]!=nil {return nil,errors.New("ambiguous or missing Level 1 assessment")}
        var w cmmc.Workspace
        if err:=json.Unmarshal(data,&w);err!=nil{return nil,err}
        if err:=cmmc.ValidatedWorkspace(w);err!=nil{return nil,err}
        if w.Level!=1{return nil,errors.New("Level 1 state disagrees with workspace")}
        for _,control:=range w.Controls {
            for _,e:=range control.Evidence {
                name:="evidence/"+e.Path
                if !snapshotIsFileName(name){return nil,errors.New("unsafe evidence reference")}
                if prior,found:=wanted[name];found && prior!=e.SHA256{return nil,errors.New("conflicting evidence hashes")}
                wanted[name]=e.SHA256
            }
        }
    }else if state.Mode=="objectives" && state.Level==2 {
        data,ok:=files["objectives.json"]
        if !ok || files["requirements.json"]!=nil{return nil,errors.New("ambiguous or missing Level 2 assessment")}
        var w cmmc.ObjectiveWorkspace
        if err:=json.Unmarshal(data,&w);err!=nil{return nil,err}
        if err:=cmmc.ValidateObjectiveWorkspace(w);err!=nil{return nil,err}
        if w.Level!=2{return nil,errors.New("Level 2 state disagrees with workspace")}
        for _,control:=range w.Controls { for _,objective:=range control.Objectives { for _,e:=range objective.Evidence {
            name:="evidence/"+e.Path
            if !snapshotIsFileName(name){return nil,errors.New("unsafe evidence reference")}
            if prior,found:=wanted[name];found && prior!=e.SHA256{return nil,errors.New("conflicting evidence hashes")}
            wanted[name]=e.SHA256
        }}}
    }else {
        return nil,errors.New("unsupported assessment mode or level")
    }
    for name,digest:=range wanted {
        data,ok:=files[name]
        if !ok{return nil,fmt.Errorf("missing referenced evidence %s",name)}
        if snapshotDigest(data)!=digest{return nil,fmt.Errorf("evidence hash mismatch: %s",name)}
    }
    for name:=range files {
        if name=="state.json" || name=="requirements.json" || name=="objectives.json" {continue}
        if _,ok:=wanted[name];!ok{return nil,fmt.Errorf("unreferenced snapshot evidence %s",name)}
    }
    return wanted,nil
}
func snapshotCollect(root string) (map[string][]byte,error) {
    abs,err:=filepath.Abs(root)
    if err!=nil{return nil,err}
    st,err:=os.Lstat(abs)
    if err!=nil{return nil,err}
    if !st.IsDir() || st.Mode()&os.ModeSymlink!=0{return nil,errors.New("unsafe assessment directory")}
    files:=map[string][]byte{}
    state,err:=snapshotReadFile(abs,"state.json",snapshotMaxFile)
    if err!=nil{return nil,err}
    files["state.json"]=state
    var s localUIState
    if err=json.Unmarshal(state,&s);err!=nil{return nil,err}
    workspace:=""
    if s.Mode=="basic" && s.Level==1 {workspace="requirements.json"}
    if s.Mode=="objectives" && s.Level==2 {workspace="objectives.json"}
    if workspace=="" {return nil,errors.New("invalid assessment mode")}
    for _,other:=range []string{"requirements.json","objectives.json"} {
        if other!=workspace {
            if _,err=os.Lstat(filepath.Join(abs,other));err==nil {
                return nil,errors.New("ambiguous assessment has both workspace types")
            }else if !os.IsNotExist(err){return nil,err}
        }
    }
    data,err:=snapshotReadFile(abs,workspace,snapshotMaxFile)
    if err!=nil{return nil,err}
    files[workspace]=data
    // Extract only locally referenced evidence; no unassociated files,
    // private access token, or unrelated user data enters the archive.
    var evidencePaths []string
    if s.Mode=="basic" {
        var w cmmc.Workspace
        if err=json.Unmarshal(data,&w);err!=nil{return nil,err}
        if err=cmmc.ValidatedWorkspace(w);err!=nil{return nil,err}
        for _,c:=range w.Controls {for _,e:=range c.Evidence {evidencePaths=append(evidencePaths,"evidence/"+e.Path)}}
    } else {
        var w cmmc.ObjectiveWorkspace
        if err=json.Unmarshal(data,&w);err!=nil{return nil,err}
        if err=cmmc.ValidateObjectiveWorkspace(w);err!=nil{return nil,err}
        for _,c:=range w.Controls {for _,o:=range c.Objectives {for _,e:=range o.Evidence {
            evidencePaths=append(evidencePaths,"evidence/"+e.Path)
        }}}
    }
    if len(evidencePaths)>snapshotMaxFiles-2{return nil,errors.New("too many evidence references")}
    total:=int64(len(state)+len(data))
    for _,name:=range evidencePaths {
        if _,ok:=files[name];ok {continue}
        blob,err:=snapshotReadFile(abs,name,snapshotMaxFile)
        if err!=nil{return nil,err}
        total+=int64(len(blob))
        if total>snapshotMaxTotal{return nil,errors.New("snapshot total size exceeds limit")}
        files[name]=blob
    }
    if _,err=snapshotRequiredEvidence(files);err!=nil{return nil,err}
    return files,nil
}
func snapshotBuildManifest(files map[string][]byte) snapshotManifest {
    m:=snapshotManifest{Schema:snapshotSchema}
    for name,b:=range files {
        m.Files=append(m.Files,snapshotEntry{Name:name,SHA256:snapshotDigest(b),Size:int64(len(b))})
    }
    sort.Slice(m.Files,func(i,j int)bool{return m.Files[i].Name<m.Files[j].Name})
    return m
}
func snapshotCreate(root,dest string) error {
    files,err:=snapshotCollect(root)
    if err!=nil{return err}
    absRoot,err:=filepath.Abs(root)
    if err!=nil{return err}
    absDest,err:=filepath.Abs(dest)
    if err!=nil{return err}
    if absDest==absRoot || strings.HasPrefix(absDest,absRoot+string(filepath.Separator)) {
        return errors.New("backup must be stored outside the live assessment directory")
    }
    f,err:=os.OpenFile(dest,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
    if err!=nil{return err}
    success:=false
    defer func(){if !success {_=os.Remove(dest)}}()
    writer:=zip.NewWriter(f)
    m:=snapshotBuildManifest(files)
    manifest,err:=json.MarshalIndent(m,"","  ")
    if err!=nil {_=writer.Close();_=f.Close();return err}
    names:=[]string{"MANIFEST.json"}
    for name:=range files {names=append(names,name)}
    sort.Strings(names)
    for _,name:=range names {
        var payload []byte
        if name=="MANIFEST.json" {payload=append(manifest,'\n')} else {payload=files[name]}
        header:=&zip.FileHeader{Name:name,Method:zip.Deflate}
        header.SetMode(0600)
        item,writeErr:=writer.CreateHeader(header)
        if writeErr!=nil {err=writeErr;break}
        _,err=item.Write(payload)
        if err!=nil {break}
    }
    if closeErr:=writer.Close();err==nil{err=closeErr}
    if syncErr:=f.Sync();err==nil{err=syncErr}
    if closeErr:=f.Close();err==nil{err=closeErr}
    if err!=nil{return err}
    success=true
    return nil
}
func snapshotLoad(path string) (map[string][]byte,error) {
    reader,err:=zip.OpenReader(path)
    if err!=nil{return nil,err}
    defer reader.Close()
    if len(reader.File)>snapshotMaxFiles+1{return nil,errors.New("backup contains too many entries")}
    files:=map[string][]byte{}
    total:=int64(0)
    for _,z:=range reader.File {
        name:=z.Name
        if name!="MANIFEST.json" && !snapshotIsFileName(name) {return nil,errors.New("unsafe entry in backup")}
        if _,exists:=files[name];exists{return nil,errors.New("duplicate backup entry")}
        if z.FileInfo().IsDir() || z.Mode()&os.ModeSymlink!=0 || !z.Mode().IsRegular() {
            return nil,errors.New("nonregular backup entry")
        }
        if z.UncompressedSize64>uint64(snapshotMaxFile){return nil,errors.New("oversized backup member")}
        rc,err:=z.Open()
        if err!=nil{return nil,err}
        b,readErr:=io.ReadAll(io.LimitReader(rc,snapshotMaxFile+1))
        closeErr:=rc.Close()
        if readErr!=nil{return nil,readErr}
        if closeErr!=nil{return nil,closeErr}
        if int64(len(b))>snapshotMaxFile || uint64(len(b))!=z.UncompressedSize64 {
            return nil,errors.New("backup member size mismatch")
        }
        total+=int64(len(b))
        if total>snapshotMaxTotal{return nil,errors.New("snapshot too large")}
        files[name]=b
    }
    blob,ok:=files["MANIFEST.json"]
    if !ok{return nil,errors.New("missing backup manifest")}
    delete(files,"MANIFEST.json")
    var manifest snapshotManifest
    if err=json.Unmarshal(blob,&manifest);err!=nil{return nil,err}
    if manifest.Schema!=snapshotSchema{return nil,errors.New("unsupported backup schema")}
    if len(manifest.Files)!=len(files){return nil,errors.New("manifest entries do not match archive")}
    seen:=map[string]bool{}
    for _,entry:=range manifest.Files {
        if seen[entry.Name]{return nil,errors.New("duplicate manifest entry")}
        seen[entry.Name]=true
        b,ok:=files[entry.Name]
        if !ok || int64(len(b))!=entry.Size || snapshotDigest(b)!=entry.SHA256{
            return nil,fmt.Errorf("backup integrity mismatch: %s",entry.Name)
        }
    }
    if _,err=snapshotRequiredEvidence(files);err!=nil{return nil,err}
    return files,nil
}
func snapshotRestore(archive,dest string) error {
    files,err:=snapshotLoad(archive)
    if err!=nil{return err}
    // Restoring into an existing directory could overwrite evidence, so a
    // freshly created destination is mandatory.
    if err=os.Mkdir(dest,0700);err!=nil{return err}
    keep:=false
    defer func(){if !keep{_=os.RemoveAll(dest)}}()
    if err=os.Mkdir(filepath.Join(dest,"evidence"),0700);err!=nil{return err}
    names:=make([]string,0,len(files))
    for name:=range files{names=append(names,name)}
    sort.Strings(names)
    for _,name:=range names {
        path:=filepath.Join(dest,filepath.FromSlash(name))
        f,err:=os.OpenFile(path,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600)
        if err!=nil{return err}
        _,writeErr:=f.Write(files[name])
        if writeErr==nil{writeErr=f.Sync()}
        closeErr:=f.Close()
        if writeErr!=nil{return writeErr}
        if closeErr!=nil{return closeErr}
    }
    if _,err=snapshotCollect(dest);err!=nil{return err}
    keep=true
    return nil
}
func cmmcBackupCommand(args []string,out io.Writer) error {
    sub:=args[0]
    fs:=flag.NewFlagSet("cmmc "+sub,flag.ContinueOnError)
    fs.SetOutput(io.Discard)
    dataDir:=fs.String("data-dir","","source local assessment directory; stop server first")
    archive:=fs.String("archive","","backup archive path (.zip)")
    dest:=fs.String("out-dir","","new restore directory (must not exist)")
    if err:=fs.Parse(args[1:]);err!=nil{return err}
    if fs.NArg()!=0{return errors.New("unexpected positional arguments")}
    if *archive==""{return errors.New("--archive required")}
    switch sub {
    case "backup":
        if *dataDir=="" || *dest!=""{return errors.New("backup requires --data-dir and --archive")}
        if err:=snapshotCreate(*dataDir,*archive);err!=nil{return err}
        fmt.Fprintln(out,"PASS: offline assessment archive created. ARCHIVE IS NOT ENCRYPTED. Store only on approved encrypted media. Stop the local UI before backup.")
        return nil
    case "verify-backup":
        if *dataDir!="" || *dest!=""{return errors.New("verify-backup requires only --archive")}
        if _,err:=snapshotLoad(*archive);err!=nil{return err}
        fmt.Fprintln(out,"PASS: snapshot hashes and workspace/evidence structure verified; this is NOT proof of CMMC compliance or encryption.")
        return nil
    case "restore-backup":
        if *dest=="" || *dataDir!=""{return errors.New("restore-backup requires --archive and new --out-dir")}
        if err:=snapshotRestore(*archive,*dest);err!=nil{return err}
        fmt.Fprintln(out,"PASS: local workspace restored to new directory, without overwriting an existing assessment. Protect plaintext files with approved encryption.")
        return nil
    default:
        return errors.New("invalid backup command")
    }
}

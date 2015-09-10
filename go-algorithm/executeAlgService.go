package main
import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strings"
 )

type SueAlg struct {
    BinaryFile string
    InputImg string
    ParamsFile string
    OutputPath string
}

type Response struct{
    IsSuccess bool `json:"isSuccess"`
    OutputPath string `json:"outputPath"`
}

func (alg SueAlg) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    
    parameters := req.URL.Query()
    log.Println("pass in query parameters:", parameters)

    //write the parameters from the request into the temp parameter file
    file, err := ioutil.TempFile(".","parameterFile")
    check(err)
    log.Println("temp parameterFileName:", file.Name())
    alg.ParamsFile = file.Name()
    
    defer os.Remove(file.Name())

    writer := bufio.NewWriter(file)
    _, err = writer.WriteString("focus :=" + parameters.Get("focus") + "\n")
    _, err = writer.WriteString("blend :=" + parameters.Get("blend") + "\n")
    _, err = writer.WriteString("edgeSmooth :=" + parameters.Get("edgeSmooth") + "\n")
    _, err = writer.WriteString("interp :=" + parameters.Get("interp") + "\n")
    _, err = writer.WriteString("first dGradThreshold :=" + parameters.Get("first dGradThreshold") + "\n")
    _, err = writer.WriteString("second dGradThreshold :=" + parameters.Get("second dGradThreshold") + "\n")
    _, err = writer.WriteString("!sharp :=" + parameters.Get("!sharp") + "\n")
    _, err = writer.WriteString("sharpen Threshold :=" + parameters.Get("sharpen Threshold") + "\n")
    _, err = writer.WriteString("gaussVar :=" + parameters.Get("gaussVar") + "\n")
    _, err = writer.WriteString("shrunkenDim :=" + parameters.Get("shrunkenDim") + "\n")
    _, err = writer.WriteString("threshold :=" + parameters.Get("threshold") + "\n")
    _, err = writer.WriteString("end of file :=" + parameters.Get("end of file") + "\n")
    
    check(err)
    writer.Flush()
    
    alg.InputImg = parameters.Get("image")+".img"

    //create temp folder for output file
    dir, err := ioutil.TempDir(".","output")
    alg.OutputPath = dir + "/"
    defer os.RemoveAll("./"+dir)

    isSuccess := RunAlgorithm(alg)

    //TODO save the output file to blob store
    
    response := Response{isSuccess, dir}
    resultJson, _ := json.Marshal(response)
    fmt.Fprint(w, string(resultJson))
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func RunAlgorithm(alg SueAlg) bool {
    fmt.Println("Start running algorithm")
    
    command := "./"+alg.BinaryFile
    args:= []string{alg.InputImg, alg.OutputPath, alg.ParamsFile}
    
    result := RunCommand(command, args)
    if result == "false" { return false }
    
    log.Println("Finish running algorithm")
    return true
}

func RunCommand(command string, args []string) string {
    runAlgrithmCmd := exec.Command(command, args...)
    PrintCommand(runAlgrithmCmd)
    output, err := runAlgrithmCmd.CombinedOutput()
    if err!= nil {
        PrintError(err)
        return "false"
    }
    PrintOutput(output)
    return (string(output))
}

func PrintCommand(cmd *exec.Cmd) {
  fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func PrintError(err error) {
  if err != nil {
    os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
  }
}

func PrintOutput(outs []byte) {
  if len(outs) > 0 {
    fmt.Printf("==> Output: %s\n", string(outs))
  }
}

func Hello(w http.ResponseWriter, r *http.Request){fmt.Fprint(w, "hello")}

//For /currentFolderContents endpoint
func RunLsCommand(w http.ResponseWriter, r *http.Request) {
    command := "ls"
    args := []string{"-al"}
    output := RunCommand(command, args)
    fmt.Fprint(w, output)
}

func main() {
    sueAlg := SueAlg{BinaryFile:"3DSUE"}
    http.Handle("/algorithms/3DSUE", &sueAlg)

    http.HandleFunc("/hello", Hello)
    http.HandleFunc("/currentFolderContents", RunLsCommand)

    var url = ":" + os.Getenv("PORT")
    
    //var url = ":8080"
    
    var err =  http.ListenAndServe(url, nil)    
    if err!= nil {log.Fatal(err)}
}
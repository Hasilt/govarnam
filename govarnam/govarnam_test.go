package govarnam

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
)

var varnamInstances = map[string]*Varnam{}
var mutex = sync.RWMutex{}
var testTempDir string

// AssertEqual checks if values are equal
// Thanks https://gist.github.com/samalba/6059502#gistcomment-2710184
func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func setUp(schemeID string, langCode string) {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := path.Join(path.Dir(filename), "..")

	vstLoc := path.Join(projectRoot, "schemes", schemeID+".vst")

	var err error
	testTempDir, err = ioutil.TempDir("", "govarnam_test")
	checkError(err)

	dictLoc := path.Join(testTempDir, langCode+".vst.learnings")

	varnam, err := Init(vstLoc, dictLoc)
	checkError(err)

	mutex.Lock()
	varnamInstances[schemeID] = varnam
	mutex.Unlock()
}

func getVarnamInstance(schemeID string) *Varnam {
	mutex.Lock()
	instance, ok := varnamInstances[schemeID]
	mutex.Unlock()

	if ok {
		return instance
	}
	return nil
}

func tearDownVarnam(schemeID string) {
	getVarnamInstance(schemeID).Close()
}

func tearDown() {
	os.RemoveAll(testTempDir)
}

func TestMain(m *testing.M) {
	schemeDetails, err := GetAllSchemeDetails()

	if err != nil {
		log.Fatal(err)
	}

	for _, schemeDetail := range schemeDetails {
		setUp(schemeDetail.Identifier, schemeDetail.LangCode)
	}

	m.Run()

	for _, schemeDetail := range schemeDetails {
		tearDownVarnam(schemeDetail.Identifier)
	}

	tearDown()
}

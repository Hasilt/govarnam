package govarnam

import (
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"runtime"
	"testing"
)

var (
	dictDir string
	varnam  Varnam
)

// AssertEqual checks if values are equal
func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func setUp(langCode string) {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := path.Join(path.Dir(filename), "..")

	vstLoc := path.Join(projectRoot, "schemes", langCode+".vst")

	dictDir, err := ioutil.TempDir("", "govarnam_test")
	if err != nil {
		log.Fatal(err)
	}

	dictLoc := path.Join(dictDir, langCode+".vst.learnings")
	makeDictionary(dictLoc)

	varnam = Init(vstLoc, dictLoc)
}

func tearDown() {
	// os.RemoveAll(dictDir)
}

func TestGreedyTokenizer(t *testing.T) {
	assertEqual(t, varnam.Transliterate("namaskaaram").GreedyTokenized[0].Word, "നമസ്കാരം")
	assertEqual(t, varnam.Transliterate("malayalam").GreedyTokenized[0].Word, "മലയലം")
}

func TestTokenizer(t *testing.T) {
	expected := []string{"മല", "മാല", "മള", "മലാ", "മളാ", "മാള", "മാലാ", "മാളാ"}
	for i, sug := range varnam.Transliterate("mala").Suggestions {
		assertEqual(t, sug.Word, expected[i])
	}
}

func TestLearn(t *testing.T) {
	assertEqual(t, varnam.Transliterate("malayalam").Suggestions[0].Word, "മലയലം")
	varnam.Learn("മലയാളം")
	assertEqual(t, varnam.Transliterate("malayalam").Suggestions[0].Word, "മലയാളം")
}

func TestMain(m *testing.M) {
	setUp("ml")
	m.Run()
	tearDown()
}
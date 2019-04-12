package parseutil

import (
	"testing"
)

//func TestParseFilePos1(t *testing.T) {
//	s := "/a/b/c:1:2"
//	fp, err := ParseFilePos(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !(fp.Filename == "/a/b/c" &&
//		fp.Line == 1 && fp.Column == 2) {
//		t.Fatal()
//	}
//}

//func TestParseFilePos2(t *testing.T) {
//	s := "/a/b\\ b/c"
//	fp, err := ParseFilePos(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if !(fp.Filename == "/a/b\\ b/c") {
//		t.Fatalf("%v", fp.Filename)
//	}
//}

//func TestParseFilePos3(t *testing.T) {
//	s := "/a/b\\"
//	fp, err := ParseFilePos(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	//if !(fp.Filename == "/a/b\\") {
//	//	t.Fatalf("%v", fp.Filename)
//	//}
//	if !(fp.Filename == "/a/b") {
//		t.Fatalf("%v", fp.Filename)
//	}
//}

//----------

//func TestExpandLastIndexOfFilename1(t *testing.T) {
//	s := ": /a/b/c"
//	i := ExpandLastIndexOfFilenameFmt(s, 100)
//	if !(i == 2) {
//		t.Fatalf("%v", i)
//	}
//}

//----------

func TestDetectVar(t *testing.T) {
	str := "aaaa$b $cd $e"
	if !DetectEnvVar(str, "b") {
		t.Fatal()
	}
	if !DetectEnvVar(str, "cd") {
		t.Fatal()
	}
	if !DetectEnvVar(str, "e") {
		t.Fatal()
	}

	str2 := "$a"
	if !DetectEnvVar(str2, "a") {
		t.Fatal()
	}
}

package difflib

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

func assertAlmostEqual(t *testing.T, a, b float64, places int) {
	if math.Abs(a-b) > math.Pow10(-places) {
		t.Errorf("%.7f != %.7f", a, b)
	}
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}

func splitChars(s string) []string {
	chars := make([]string, 0, len(s))
	// Assume ASCII inputs
	for i := 0; i != len(s); i++ {
		chars = append(chars, string(s[i]))
	}
	return chars
}

func TestSequenceMatcherRatio(t *testing.T) {
	s := NewMatcher(splitChars("abcd"), splitChars("bcde"))
	assertEqual(t, s.Ratio(), 0.75)
	assertEqual(t, s.QuickRatio(), 0.75)
	assertEqual(t, s.RealQuickRatio(), 1.0)
}

func TestGetOptCodes(t *testing.T) {
	a := "qabxcd"
	b := "abycdf"
	s := NewMatcher(splitChars(a), splitChars(b))
	w := &bytes.Buffer{}
	for _, op := range s.GetOpCodes() {
		fmt.Fprintf(w, "%s a[%d:%d], (%s) b[%d:%d] (%s)\n", string(op.Tag),
			op.I1, op.I2, a[op.I1:op.I2], op.J1, op.J2, b[op.J1:op.J2])
	}
	result := w.String()
	expected := `d a[0:1], (q) b[0:0] ()
e a[1:3], (ab) b[0:2] (ab)
r a[3:4], (x) b[2:3] (y)
e a[4:6], (cd) b[3:5] (cd)
i a[6:6], () b[5:6] (f)
`
	if expected != result {
		t.Errorf("unexpected op codes: \n%s", result)
	}
}

func TestGroupedOpCodes(t *testing.T) {
	a := []string{}
	for i := 0; i != 39; i++ {
		a = append(a, fmt.Sprintf("%02d", i))
	}
	b := []string{}
	b = append(b, a[:8]...)
	b = append(b, " i")
	b = append(b, a[8:19]...)
	b = append(b, " x")
	b = append(b, a[20:22]...)
	b = append(b, a[27:34]...)
	b = append(b, " y")
	b = append(b, a[35:]...)
	s := NewMatcher(a, b)
	w := &bytes.Buffer{}
	for _, g := range s.GetGroupedOpCodes(-1) {
		fmt.Fprintf(w, "group\n")
		for _, op := range g {
			fmt.Fprintf(w, "  %s, %d, %d, %d, %d\n", string(op.Tag),
				op.I1, op.I2, op.J1, op.J2)
		}
	}
	result := w.String()
	expected := `group
  e, 5, 8, 5, 8
  i, 8, 8, 8, 9
  e, 8, 11, 9, 12
group
  e, 16, 19, 17, 20
  r, 19, 20, 20, 21
  e, 20, 22, 21, 23
  d, 22, 27, 23, 23
  e, 27, 30, 23, 26
group
  e, 31, 34, 27, 30
  r, 34, 35, 30, 31
  e, 35, 38, 31, 34
`
	if expected != result {
		t.Errorf("unexpected op codes: \n%s", result)
	}
}

func ExampleGetUnifiedDiffStringCode() {
	a := `one
two
three
four
fmt.Printf("%s,%T",a,b)`
	b := `zero
one
three
four`
	diff := UnifiedDiff{
		A:        SplitLines(a),
		B:        SplitLines(b),
		FromFile: "Original",
		FromDate: "2005-01-26 23:30:50",
		ToFile:   "Current",
		ToDate:   "2010-04-02 10:20:52",
		Context:  3,
	}
	result, _ := GetUnifiedDiffString(diff)
	fmt.Println(strings.Replace(result, "\t", " ", -1))
	// Output:
	// --- a/Original 2005-01-26 23:30:50
	// +++ b/Current 2010-04-02 10:20:52
	// @@ -1,5 +1,4 @@
	// +zero
	//  one
	// -two
	//  three
	//  four
	// -fmt.Printf("%s,%T",a,b)
}

func ExampleGetContextDiffStringCode() {
	a := `one
two
three
four
fmt.Printf("%s,%T",a,b)`
	b := `zero
one
tree
four`
	diff := ContextDiff{
		A:        SplitLines(a),
		B:        SplitLines(b),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
		Eol:      "\n",
	}
	result, _ := GetContextDiffString(diff)
	fmt.Print(strings.Replace(result, "\t", " ", -1))
	// Output:
	// *** Original
	// --- Current
	// ***************
	// *** 1,5 ****
	//   one
	// ! two
	// ! three
	//   four
	// - fmt.Printf("%s,%T",a,b)
	// --- 1,4 ----
	// + zero
	//   one
	// ! tree
	//   four
}

func ExampleGetContextDiffString() {
	a := `one
two
three
four`
	b := `zero
one
tree
four`
	diff := ContextDiff{
		A:        SplitLines(a),
		B:        SplitLines(b),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
		Eol:      "\n",
	}
	result, _ := GetContextDiffString(diff)
	fmt.Print(strings.Replace(result, "\t", " ", -1))
	// Output:
	// *** Original
	// --- Current
	// ***************
	// *** 1,4 ****
	//   one
	// ! two
	// ! three
	//   four
	// --- 1,4 ----
	// + zero
	//   one
	// ! tree
	//   four
}

func rep(s string, count int) string {
	return strings.Repeat(s, count)
}

func TestWithAsciiOneInsert(t *testing.T) {
	sm := NewMatcher(splitChars(rep("b", 100)),
		splitChars("a"+rep("b", 100)))
	assertAlmostEqual(t, sm.Ratio(), 0.995, 3)
	assertEqual(t, sm.GetOpCodes(),
		[]OpCode{{'i', 0, 0, 0, 1}, {'e', 0, 100, 1, 101}})
	assertEqual(t, len(sm.bPopular), 0)

	sm = NewMatcher(splitChars(rep("b", 100)),
		splitChars(rep("b", 50)+"a"+rep("b", 50)))
	assertAlmostEqual(t, sm.Ratio(), 0.995, 3)
	assertEqual(t, sm.GetOpCodes(),
		[]OpCode{{'e', 0, 50, 0, 50}, {'i', 50, 50, 50, 51}, {'e', 50, 100, 51, 101}})
	assertEqual(t, len(sm.bPopular), 0)
}

func TestWithAsciiOnDelete(t *testing.T) {
	sm := NewMatcher(splitChars(rep("a", 40)+"c"+rep("b", 40)),
		splitChars(rep("a", 40)+rep("b", 40)))
	assertAlmostEqual(t, sm.Ratio(), 0.994, 3)
	assertEqual(t, sm.GetOpCodes(),
		[]OpCode{{'e', 0, 40, 0, 40}, {'d', 40, 41, 40, 40}, {'e', 41, 81, 40, 80}})
}

func TestWithAsciiBJunk(t *testing.T) {
	isJunk := func(s string) bool {
		return s == " "
	}
	sm := NewMatcherWithJunk(splitChars(rep("a", 40)+rep("b", 40)),
		splitChars(rep("a", 44)+rep("b", 40)), true, isJunk)
	assertEqual(t, sm.bJunk, map[string]struct{}{})

	sm = NewMatcherWithJunk(splitChars(rep("a", 40)+rep("b", 40)),
		splitChars(rep("a", 44)+rep("b", 40)+rep(" ", 20)), false, isJunk)
	assertEqual(t, sm.bJunk, map[string]struct{}{" ": {}})

	isJunk = func(s string) bool {
		return s == " " || s == "b"
	}
	sm = NewMatcherWithJunk(splitChars(rep("a", 40)+rep("b", 40)),
		splitChars(rep("a", 44)+rep("b", 40)+rep(" ", 20)), false, isJunk)
	assertEqual(t, sm.bJunk, map[string]struct{}{" ": {}, "b": {}})
}

func TestSFBugsRatioForNullSeqn(t *testing.T) {
	sm := NewMatcher(nil, nil)
	assertEqual(t, sm.Ratio(), 1.0)
	assertEqual(t, sm.QuickRatio(), 1.0)
	assertEqual(t, sm.RealQuickRatio(), 1.0)
}

func TestSFBugsComparingEmptyLists(t *testing.T) {
	groups := NewMatcher(nil, nil).GetGroupedOpCodes(-1)
	assertEqual(t, len(groups), 0)
	diff := UnifiedDiff{
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
	}
	result, err := GetUnifiedDiffString(diff)
	assertEqual(t, err, nil)
	assertEqual(t, result, "")
}

func TestOutputFormatRangeFormatUnified(t *testing.T) {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	//
	// Each <range> field shall be of the form:
	//   %1d", <beginning line number>  if the range contains exactly one line,
	// and:
	//  "%1d,%1d", <beginning line number>, <number of lines> otherwise.
	// If a range is empty, its beginning line number shall be the number of
	// the line just before the range, or 0 if the empty range starts the file.
	fm := formatRangeUnified
	assertEqual(t, fm(3, 3), "3,0")
	assertEqual(t, fm(3, 4), "4")
	assertEqual(t, fm(3, 5), "4,2")
	assertEqual(t, fm(3, 6), "4,3")
	assertEqual(t, fm(0, 0), "0,0")
}

func TestOutputFormatRangeFormatContext(t *testing.T) {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	//
	// The range of lines in file1 shall be written in the following format
	// if the range contains two or more lines:
	//     "*** %d,%d ****\n", <beginning line number>, <ending line number>
	// and the following format otherwise:
	//     "*** %d ****\n", <ending line number>
	// The ending line number of an empty range shall be the number of the preceding line,
	// or 0 if the range is at the start of the file.
	//
	// Next, the range of lines in file2 shall be written in the following format
	// if the range contains two or more lines:
	//     "--- %d,%d ----\n", <beginning line number>, <ending line number>
	// and the following format otherwise:
	//     "--- %d ----\n", <ending line number>
	fm := formatRangeContext
	assertEqual(t, fm(3, 3), "3")
	assertEqual(t, fm(3, 4), "4")
	assertEqual(t, fm(3, 5), "4,5")
	assertEqual(t, fm(3, 6), "4,6")
	assertEqual(t, fm(0, 0), "0")
}

func TestOutputFormatTabDelimiter(t *testing.T) {
	diff := UnifiedDiff{
		A:        splitChars("one"),
		B:        splitChars("two"),
		FromFile: "Original",
		FromDate: "2005-01-26 23:30:50",
		ToFile:   "Current",
		ToDate:   "2010-04-12 10:20:52",
		Eol:      "\n",
	}
	ud, err := GetUnifiedDiffString(diff)
	assertEqual(t, err, nil)
	assertEqual(t, SplitLines(ud)[:2], []string{
		"--- a/Original\t2005-01-26 23:30:50\n",
		"+++ b/Current\t2010-04-12 10:20:52\n",
	})
	cd, err := GetContextDiffString(ContextDiff(diff))
	assertEqual(t, err, nil)
	assertEqual(t, SplitLines(cd)[:2], []string{
		"*** Original\t2005-01-26 23:30:50\n",
		"--- Current\t2010-04-12 10:20:52\n",
	})
}

func TestOutputFormatNoTrailingTabOnEmptyFiledate(t *testing.T) {
	diff := UnifiedDiff{
		A:        splitChars("one"),
		B:        splitChars("two"),
		FromFile: "Original",
		ToFile:   "Current",
		Eol:      "\n",
	}
	ud, err := GetUnifiedDiffString(diff)
	assertEqual(t, err, nil)
	assertEqual(t, SplitLines(ud)[:2], []string{"--- a/Original\n", "+++ b/Current\n"})

	cd, err := GetContextDiffString(ContextDiff(diff))
	assertEqual(t, err, nil)
	assertEqual(t, SplitLines(cd)[:2], []string{"*** Original\n", "--- Current\n"})
}

func TestOmitFilenames(t *testing.T) {
	diff := UnifiedDiff{
		A:   SplitLines("o\nn\ne\n"),
		B:   SplitLines("t\nw\no\n"),
		Eol: "\n",
	}
	ud, err := GetUnifiedDiffString(diff)
	assertEqual(t, err, nil)
	assertEqual(t, SplitLines(ud), []string{
		"--- a/\n",
		"+++ b/\n",
		"@@ -0,0 +1,2 @@\n",
		"+t\n",
		"+w\n",
		"@@ -2,2 +3,0 @@ o\n",
		"-n\n",
		"-e\n",
		"\n",
	})

	cd, err := GetContextDiffString(ContextDiff(diff))
	assertEqual(t, err, nil)
	assertEqual(t, SplitLines(cd), []string{
		"***************\n",
		"*** 0 ****\n",
		"--- 1,2 ----\n",
		"+ t\n",
		"+ w\n",
		"***************\n",
		"*** 2,3 ****\n",
		"- n\n",
		"- e\n",
		"--- 3 ----\n",
		"\n",
	})
}

func TestSplitLines(t *testing.T) {
	allTests := []struct {
		input string
		want  []string
	}{
		{"foo", []string{"foo\n"}},
		{"foo\nbar", []string{"foo\n", "bar\n"}},
		{"foo\nbar\n", []string{"foo\n", "bar\n", "\n"}},
	}
	for _, test := range allTests {
		assertEqual(t, SplitLines(test.input), test.want)
	}
}

func TestUnifiedDiffWithContext(t *testing.T) {
	a := []string{
		"class Example:\n",
		"    def __init__(self):\n",
		"        self.x = 1\n",
		"        self.y = 2\n",
		"    def method(self):\n",
		"        return self.x + self.y\n",
	}
	b := []string{
		"class Example:\n",
		"    def __init__(self):\n",
		"        self.x = 1\n",
		"        self.z = 3\n",
		"    def method(self):\n",
		"        return self.x + self.z\n",
	}

	diff := UnifiedDiff{
		A:        a,
		B:        b,
		FromFile: "original.py",
		ToFile:   "modified.py",
		Context:  3,
	}

	result, err := GetUnifiedDiffString(diff)
	assertEqual(t, err, nil)

	// Split the result into lines for easier comparison
	lines := strings.Split(result, "\n")

	// Find the @@ line
	var foundHeader bool
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			foundHeader = true
			break
		}
	}
	assertEqual(t, foundHeader, true)
}

func TestUnifiedDiffWithLongContext(t *testing.T) {
	a := []string{
		"This is a very long line that should be truncated in the context................................................\n",
		"line0\n",
		"line1\n",
		"line2\n",
		"line3\n",
	}
	b := []string{
		"This is a very long line that should be truncated in the context................................................\n",
		"line0\n",
		"line1\n",
		"line2\n",
		"modified line3\n",
	}

	diff := UnifiedDiff{
		A:        a,
		B:        b,
		FromFile: "original.txt",
		ToFile:   "modified.txt",
		Context:  3,
	}

	result, err := GetUnifiedDiffString(diff)
	assertEqual(t, err, nil)

	lines := strings.Split(result, "\n")

	// Find the @@ line
	var headerLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			headerLine = line
			break
		}
	}

	// // Verify that the context is truncated with ...
	assertEqual(t, len(headerLine)-len("@@ -2,4 +2,4 @@ ") <= 80, true)
}

func TestUnifiedDiffCodeWithContext(t *testing.T) {
	a := []string{
		"package main\n",
		"\n",
		"import \"fmt\"\n",
		"\n",
		"func main() {\n",
		"    fmt.Println(\"test\")\n",
		"}\n",
	}
	b := []string{
		"package main\n",
		"\n",
		"import \"fmt\"\n",
		"\n",
		"func main() {\n",
		"    fmt.Println(\"test\")\n",
		"    fmt.Println(\"kk\")\n",
		"}\n",
	}

	diff := UnifiedDiff{
		A:        a,
		B:        b,
		FromFile: "main.go",
		ToFile:   "main.go",
		Context:  3,
	}

	result, err := GetUnifiedDiffString(diff)
	assertEqual(t, err, nil)

	expectedDiff := `--- a/main.go
+++ b/main.go
@@ -4,4 +4,5 @@ import "fmt"
 
 func main() {
     fmt.Println("test")
+    fmt.Println("kk")
 }
`

	compareByteByByte(t, normalizeLineEndings(result), normalizeLineEndings(expectedDiff))
	assertEqual(t, normalizeLineEndings(result), normalizeLineEndings(expectedDiff))
}

func compareByteByByte(t *testing.T, got, want string) {
	got = normalizeLineEndings(got)
	want = normalizeLineEndings(want)

	// 先比较长度
	if len(got) != len(want) {
		t.Errorf("Length mismatch: got %d bytes, want %d bytes", len(got), len(want))
	}

	// 转换为字节数组进行对比
	gotBytes := []byte(got)
	wantBytes := []byte(want)

	// 找出第一个不同的位置
	var diffPos int
	for diffPos = 0; diffPos < len(gotBytes) && diffPos < len(wantBytes); diffPos++ {
		if gotBytes[diffPos] != wantBytes[diffPos] {
			break
		}
	}

	if diffPos < len(gotBytes) || diffPos < len(wantBytes) {
		// 计算显示的上下文范围
		start := diffPos - 20
		if start < 0 {
			start = 0
		}
		endGot := diffPos + 20
		if endGot > len(gotBytes) {
			endGot = len(gotBytes)
		}
		endWant := diffPos + 20
		if endWant > len(wantBytes) {
			endWant = len(wantBytes)
		}

		// 构建详细的错误信息
		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("Difference at position %d:\n", diffPos))

		// 显示字节值
		if diffPos < len(gotBytes) {
			msg.WriteString(fmt.Sprintf("Got byte: %d (%q)\n", gotBytes[diffPos], string(gotBytes[diffPos])))
		}
		if diffPos < len(wantBytes) {
			msg.WriteString(fmt.Sprintf("Want byte: %d (%q)\n", wantBytes[diffPos], string(wantBytes[diffPos])))
		}

		// 显示上下文
		msg.WriteString("\nContext:\n")
		msg.WriteString("Got:  ")
		for i := start; i < endGot; i++ {
			if i == diffPos {
				msg.WriteString("[" + string(gotBytes[i]) + "]")
			} else {
				msg.WriteString(string(gotBytes[i]))
			}
		}
		msg.WriteString("\nWant: ")
		for i := start; i < endWant; i++ {
			if i == diffPos {
				msg.WriteString("[" + string(wantBytes[i]) + "]")
			} else {
				msg.WriteString(string(wantBytes[i]))
			}
		}

		// 输出16进制表示
		msg.WriteString("\n\nHex dump:\n")
		msg.WriteString("Got:  ")
		for i := start; i < endGot; i++ {
			if i == diffPos {
				msg.WriteString("[" + fmt.Sprintf("%02x", gotBytes[i]) + "]")
			} else {
				msg.WriteString(fmt.Sprintf("%02x", gotBytes[i]))
			}
			msg.WriteString(" ")
		}
		msg.WriteString("\nWant: ")
		for i := start; i < endWant; i++ {
			if i == diffPos {
				msg.WriteString("[" + fmt.Sprintf("%02x", wantBytes[i]) + "]")
			} else {
				msg.WriteString(fmt.Sprintf("%02x", wantBytes[i]))
			}
			msg.WriteString(" ")
		}

		t.Error("\n" + msg.String())
	}
}

func normalizeLineEndings(s string) string {
	// First replace Windows line endings with Unix ones
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// Then replace any remaining lone CR with LF
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func benchmarkSplitLines(b *testing.B, count int) {
	str := strings.Repeat("foo\n", count)

	b.ResetTimer()

	n := 0
	for i := 0; i < b.N; i++ {
		n += len(SplitLines(str))
	}
}

func BenchmarkSplitLines100(b *testing.B) {
	benchmarkSplitLines(b, 100)
}

func BenchmarkSplitLines10000(b *testing.B) {
	benchmarkSplitLines(b, 10000)
}

package ow

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestOwBasic(t *testing.T) {
	n1 := runtime.NumGoroutine()

	expected := []string{
		"c1-1",
		"c1-2",
		"c1-3",
		"c2-1",
		"c2-2",
		"c3-1",
		"c3-2",
	}

	buf := bytes.Buffer{}
	o := New(&buf)
	n2 := runtime.NumGoroutine()
	if n1+1 != n2 {
		t.Errorf("NumGoroutine %d %d", n1, n2)
	}

	c1 := o.GetW()
	fmt.Fprintf(c1, "c1-1")
	fmt.Fprintf(c1, "c1-2")

	c2 := o.GetW()

	c3 := o.GetW()
	fmt.Fprintf(c3, "c3-1")
	fmt.Fprintf(c2, "c2-1")
	fmt.Fprintf(c3, "c3-2")
	fmt.Fprintf(c2, "c2-2")
	fmt.Fprintf(c1, "c1-3")

	c1.Close()
	c2.Close()
	c3.Close()

	n3 := runtime.NumGoroutine()
	if n2 != n3 {
		t.Errorf("NumGoroutine %d %d", n2, n3)
	}

	err := o.Wait()
	if err != nil {
		t.Fatal(err)
	}

	if g, e := buf.String(), strings.Join(expected, ``); g != e {
		t.Errorf("got %q, want %q", g, e)
	}

	n4 := runtime.NumGoroutine()
	if n1 != n4 {
		t.Errorf("NumGoroutine %d %d", n1, n4)
	}
}

func ExampleOw() {
	buf := bytes.Buffer{}
	o := New(&buf)

	w1 := o.GetW()
	w2 := o.GetW()
	w3 := o.GetW()

	time.Sleep(1 * time.Millisecond)

	fmt.Fprintln(w1, "Hello c1")
	fmt.Fprintln(w2, "Hello c2")
	fmt.Fprintln(w3, "Hello c3")
	fmt.Fprintln(w3, "Hello c3 again")
	fmt.Fprintln(w1, "Hello c1 again")
	w1.Close()
	fmt.Fprintln(w2, "Hello c2 again")
	fmt.Fprintln(w1, "Bye c1")
	fmt.Fprintln(w2, "Bye c2")

	w2.Close()
	fmt.Fprintln(w3, "Bye c3")
	w3.Close()

	o.Wait()

	fmt.Println(buf.String())
	// Output:
	// Hello c1
	// Hello c1 again
	// Bye c1
	// Hello c2
	// Hello c2 again
	// Bye c2
	// Hello c3
	// Hello c3 again
	// Bye c3
}

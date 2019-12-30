[![GoDoc Reference](https://godoc.org/github.com/sago35/ow?status.svg)](https://godoc.org/github.com/sago35/ow)

# Ow

Package ow provides ordered io.Writer.

## Usage

```go
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
```

## Licence

MIT

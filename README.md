# simple-png
pure go png parse

```go

	open, err := os.Open("./demo.png")
	if err != nil {
		panic(err)
	}
	p, err := ParsePng(open)
	if err != nil {
		panic(err)
	}
	// print png image data
	log.Println(*p.IDATs[0])
	// print png addition text
	log.Println(*p.TEXTs[0])
	
	// print other chunks
	for i := range p.chunks {
		log.Println(string(p.chunks[i].code[:]))
	}
```
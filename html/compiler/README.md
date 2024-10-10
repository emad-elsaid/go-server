to generate tags

from the toplevel dir 

```bash
go run compiler.go -input ../tag/tags.txt -output ../tag/tags.go -pkg tag -template tag
```

for attributes

```bash
go run compiler.go -input ../attr/attrs.txt -output ../attr/attrs.go -pkg attr -template attr
```

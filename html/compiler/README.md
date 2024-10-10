to generate tags

from the toplevel dir 

```bash
go run compile.go -input ../tag/tags.txt -output ../tag/tags.go -pkg tag -template tag
```

for attributes

```bash
go run compile.go -input ../attr/attrs.txt -output ../attr/attrs.go -pkg attr -template attr
```

for void tags

```bash
go run compile.go -input ../tag/voidtags.txt -output ../tag/voidtags.go -pkg tag -template voidtag
```

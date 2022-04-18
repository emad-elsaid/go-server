package main

func main() {
	GET("/", func(w Response, r Request) Output {
		return Render("layout", "index", Locals{"csrf": CSRF(r)})
	})

	Helpers()
	Start()
}

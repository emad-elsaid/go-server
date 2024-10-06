package main

func main() {
	Get("/", func(w Response, r Request) Output {
		return Render("layout", "index", Locals{"csrf": CSRF(r)})
	})

	Start()
}

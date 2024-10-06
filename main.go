package main

func main() {
	Get("/", func(r Request) Output {
		return Render("layout", "index", Locals{"csrf": CSRF(r)})
	})

	Start()
}

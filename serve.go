package gospel

func Serve(app App) error {
	server := MakeServer(app)
	return server.Start()
}

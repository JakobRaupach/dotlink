package dotlink

func (app *appEnv) reload() error {
	err := app.delete()
	if err != nil {
		return err
	}
	return app.link()
}


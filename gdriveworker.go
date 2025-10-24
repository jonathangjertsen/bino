package main

type GDriveWorker struct {
	cfg GDriveConfig
	g   *GDrive
	c   *Cache
}

func NewGDriveWorker(cfg GDriveConfig, g *GDrive, c *Cache) *GDriveWorker {
	w := &GDriveWorker{
		cfg: cfg,
		g:   g,
		c:   c,
	}

	// Warm the cache on gdrive info
	go func() {
		c.Delete("gdrive-config-info")
		_ = w.GetGDriveConfigInfo()
	}()

	return w
}

type GDriveConfigInfo struct {
	JournalFolder GDriveItem
	TemplateFile  GDriveItem
}

func (w *GDriveWorker) GetGDriveConfigInfo() GDriveConfigInfo {
	var configInfo GDriveConfigInfo
	w.c.Cached("gdrive-config-info", &configInfo, func() error {
		item, err := w.g.GetFile(w.cfg.JournalFolder)
		if err != nil {
			return err
		}
		configInfo.JournalFolder = item

		item, err = w.g.GetFile(w.cfg.TemplateFile)
		if err != nil {
			return err
		}
		configInfo.TemplateFile = item

		return nil
	})
	return configInfo
}

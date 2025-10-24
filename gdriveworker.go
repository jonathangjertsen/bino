package main

import (
	"log"
)

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
	TemplateDoc   GDriveJournal
}

func (w *GDriveWorker) GetGDriveConfigInfo() GDriveConfigInfo {
	var configInfo GDriveConfigInfo
	w.c.Cached("gdrive-config-info", &configInfo, func() error {
		item, err := w.g.GetFile(w.cfg.JournalFolder)
		if err != nil {
			return err
		}
		configInfo.JournalFolder = item

		doc, err := w.g.ReadDocument(w.cfg.TemplateFile)
		if err != nil {
			return err
		}
		configInfo.TemplateDoc = doc

		if err := doc.Validate(); err != nil {
			return err
		}
		log.Printf("Fetched GDrive Config info")

		return nil
	})
	return configInfo
}

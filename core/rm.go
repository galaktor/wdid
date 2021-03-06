package core

import (
	"context"
)

func Rm(ctx context.Context, idString string) error {
	store := ctx.Value("store").(Store)
	item, err := store.Find(idString)
	if err != nil {
		return err
	}
	return store.Delete(item)
}

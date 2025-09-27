package cmd

import (
	"io"

	"github.com/pkg/errors"
	"github.com/tobischo/gokeepasslib/v3"
)

func openKbdx(rd io.Reader, credentials *gokeepasslib.DBCredentials) (*gokeepasslib.Database, error) {

	db := gokeepasslib.NewDatabase()
	db.Credentials = credentials

	err := gokeepasslib.NewDecoder(rd).Decode(db)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = db.UnlockProtectedEntries()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return db, nil
}

func getCred(passwd string, keyfilepath string) *gokeepasslib.DBCredentials {
	var cred *gokeepasslib.DBCredentials

	if passwd != "" {
		if keyfilepath != "" {
			cred, _ = gokeepasslib.NewPasswordAndKeyCredentials(passwd, keyfilepath)
		} else {
			cred = gokeepasslib.NewPasswordCredentials(passwd)
		}
	} else {
		if keyfilepath != "" {
			cred, _ = gokeepasslib.NewKeyCredentials(keyfilepath)
		} else {
			panic("Your must either configure `keepassxc_master_password` or `keepassxc_keyfile_path`")
		}
	}
	return cred
}

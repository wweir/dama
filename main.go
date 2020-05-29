package main

import (
	"net/http"

	"github.com/wweir/dama/sftp"
	"github.com/wweir/utils/log"
	"golang.org/x/net/webdav"
)

func main() {
	// sshInfo, err := sftp.NewSftpDriver("tx.wweir.cc", "root", "", "")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(sshInfo.Stat(context.Background(), "/"))

	dav()
}

func dav() {
	sshInfo, err := sftp.NewSftpDriver("tx.wweir.cc", "root", "", "")
	if err != nil {
		panic(err)
	}
	log.Infow("ssh", "info", sshInfo)

	davHandler := &webdav.Handler{
		FileSystem: sshInfo,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			log.Infow("sftp", "methed", r.Method, "req", sshInfo.HomePath(r.URL.String()), "err", err)
		},
	}
	panic(http.ListenAndServe(":8888", davHandler))
}

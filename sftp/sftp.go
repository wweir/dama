package sftp

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// GetDefaultSSHKeyPath default private key path
func GetDefaultSSHKeyPath() string {
	home, _ := os.UserHomeDir()
	fs, err := ioutil.ReadDir(filepath.Join(home, "/.ssh"))
	if err != nil {
		return ""
	}

	keys := make([]string, 0, 1)
	for _, f := range fs {
		if strings.HasPrefix(f.Name(), "id_") && !strings.HasSuffix(f.Name(), ".pub") {
			keys = append(keys, home+"/.ssh/"+f.Name())
		}
	}
	if len(keys) == 1 {
		return keys[0]
	}
	return ""
}

// SSHInfo is a wrap for sftp client for webdav
type SSHInfo struct {
	IP       string
	User     string
	Key      string
	Password string
	Home     string

	ssh      *ssh.Client
	sftp     *sftp.Client
	dirCache *sync.Map
}

func NewSftpDriver(info, defaultUser, defaultKey, defaultPasswd string) (sshInfo *SSHInfo, err error) {
	sshInfo = &SSHInfo{
		IP:       info,
		User:     defaultUser,
		Key:      defaultKey,
		Password: defaultPasswd,
		dirCache: &sync.Map{},
	}
	if sshInfo.User == "" {
		sshInfo.User = os.Getenv("USER")
	}
	if sshInfo.Key == "" {
		sshInfo.Key = GetDefaultSSHKeyPath()
	}

	if idxIP := strings.LastIndex(info, "@"); idxIP >= 0 {
		sshInfo.IP = info[idxIP+1:]

		if idxUser := strings.LastIndex(info[:idxIP], ":"); idxUser >= 0 {
			sshInfo.User = info[:idxUser]
			sshInfo.Password = info[idxUser+1 : idxIP]
		} else {
			sshInfo.User = info[:idxIP]
		}
	}

	if sshInfo.ssh, sshInfo.sftp, err = sshInfo.Dial(); err != nil {
		return nil, err
	}

	sess, err := sshInfo.ssh.NewSession()
	if err != nil {
		return nil, err
	}
	out, err := sess.CombinedOutput("echo $PWD")
	if err != nil {
		return nil, err
	}
	sshInfo.Home = strings.TrimSuffix(string(out), "\n")

	return sshInfo, nil
}

// Dial build a wrapped ssh client with given config
func (s *SSHInfo) Dial() (*ssh.Client, *sftp.Client, error) {
	var auth ssh.AuthMethod
	if s.Password != "" {
		auth = ssh.Password(s.Password)

	} else {
		keyData, err := ioutil.ReadFile(s.Key)
		if err != nil {
			return nil, nil, err
		}
		privateKey, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, nil, err
		}
		auth = ssh.PublicKeys(privateKey)
	}

	config := &ssh.ClientConfig{
		User:            s.User,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: func(string, net.Addr, ssh.PublicKey) error { return nil },
	}

	ip := s.IP
	if !strings.Contains(s.IP, ":") {
		ip += ":22"
	}

	sshClient, err := ssh.Dial("tcp", ip, config)
	if err != nil {
		return nil, nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, nil, err
	}

	return sshClient, sftpClient, nil
}

package sshdocker

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/kohkimakimoto/loglv"
	"github.com/kr/pty"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

type Server struct {
	Config *Config
}

func NewServer(config ...*Config) *Server {
	var c *Config
	if len(config) == 0 {
		c = NewConfig()
	} else {
		c = config[0]
	}

	return &Server{
		Config: c,
	}
}

func (srv *Server) Run() error {
	config := srv.Config

	if loglv.IsDebug() {
		log.Print("[DEBUG] boot server")
	}

	ssh.Handle(func(sess ssh.Session) {
		user := sess.User()

		runtime := srv.runtime(user)
		if runtime == nil {
			log.Printf("Unsupported runtime: '%s'\n", user)
			io.WriteString(sess, fmt.Sprintf("Unsupported runtime: '%s'\n", user))
			sess.Exit(1)
			return
		}

		log.Printf("creating container with runtime: '%s'\n", runtime.Name)

		status, err := srv.runDockerContainer(sess, runtime)
		if err != nil {
			io.WriteString(sess, fmt.Sprintf("[error] %v\n", err))
			log.Printf("[error] %v", err)
		}

		sess.Exit(status)

		if loglv.IsDebug() {
			log.Printf("[DEBUG] ssh connection exited with status %d", status)
		}
	})

	var options []ssh.Option

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		if loglv.IsDebug() {
			log.Printf("[DEBUG] reloading config")
		}

		if err := config.Reload(); err != nil {
			log.Printf("[error] failed to reload config: %v", err)
		}

		if loglv.IsDebug() {
			log.Printf("[DEBUG] validating public key")
		}

		publicKeyAuthentication := config.PublicKeyAuthentication
		authorizedKeysPath := config.AuthorizedKeysFile
		authorizedKeys := config.AuthorizedKeys

		// override config if the runtime has configs.
		runtime := srv.runtime(ctx.User())
		if runtime != nil {
			if runtime.PublicKeyAuthentication != nil {
				publicKeyAuthentication = *runtime.PublicKeyAuthentication
			}

			if runtime.AuthorizedKeysFile != nil {
				authorizedKeysPath = *runtime.AuthorizedKeysFile
			}

			if runtime.AuthorizedKeys != nil {
				authorizedKeys = *runtime.AuthorizedKeys
			}
		}

		if !publicKeyAuthentication {
			return true
		}

		var keysdata []byte
		if authorizedKeysPath != "" {
			data, err := ioutil.ReadFile(authorizedKeysPath)
			if err != nil {
				log.Printf("[error] %v", err)
				return false
			}

			keysdata = data
		}

		for _, s := range authorizedKeys {
			keysdata = append(keysdata, []byte(s+"\n")...)
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(keysdata))
		for scanner.Scan() {
			keyLine := scanner.Bytes()
			if len(keyLine) == 0 {
				continue
			}

			allowed, comment, _, _, err := ssh.ParseAuthorizedKey(keyLine)
			if err != nil {
				log.Printf("[error] %v", err)
				return false
			}

			ok := ssh.KeysEqual(key, allowed)
			if ok {
				if loglv.IsDebug() {
					log.Printf("[DEBUG] authkey comment is '%s'", comment)
				}
				return true
			}

		}

		return false
	})

	options = append(options, publicKeyOption)

	if config.HostKeyFile != "" {
		if loglv.IsDebug() {
			log.Printf("[DEBUG] using host key fileis '%s'", config.HostKeyFile)
		}

		hostKeyOption := ssh.HostKeyFile(config.HostKeyFile)
		options = append(options, hostKeyOption)
	}

	log.Printf("starting ssh server on %s", config.Addr)

	return ssh.ListenAndServe(config.Addr, nil, options...)
}

func (srv *Server) runDockerContainer(sess ssh.Session, runtime *RuntimeConfig) (int, error) {
	user := sess.User()
	config := srv.Config

	ptyReq, winCh, isPty := sess.Pty()

	var args []string

	cCfg := runtime.Container
	if cCfg != nil {
		if runtime.Image != "" {
			return 1, fmt.Errorf("The runtime '%s' couldn't set image with containr. You must modify the config.", runtime.Name)
		}

		// run a docker container in the background and exec a command in the container.
		cRawImage := cCfg.Image
		cRawOptions := cCfg.Options
		cRawCommand := cCfg.Command

		// expand variables
		cImage := expand(cRawImage, user)
		var cOptions []string
		var cCommand []string
		for _, v := range cRawOptions {
			cOptions = append(cOptions, expand(v, user))
		}
		for _, v := range cRawCommand {
			cCommand = append(cCommand, expand(v, user))
		}

		var cArgs []string
		cArgs = append(cArgs, "run", "--rm", "-l", config.ContainerLabel, "-d")
		cArgs = append(cArgs, cOptions...)
		cArgs = append(cArgs, cImage)
		cArgs = append(cArgs, cCommand...)

		if loglv.IsDebug() {
			log.Printf("[DEBUG] running a contaienr in the background. docker command args: %s", cArgs)
		}

		cmd := exec.Command("docker", cArgs...)
		cmd.Stderr = sess.Stderr()
		b, err := cmd.Output()
		if err != nil {
			return exitStatus(err), err
		}

		cID := strings.TrimSpace(string(b))
		defer func() {
			if loglv.IsDebug() {
				log.Printf("[DEBUG] killing the background contaienr: %s", cID)
			}

			cmd := exec.Command("docker", "kill", cID)
			err := cmd.Run()
			if err != nil {
				log.Printf("[error] %v", err)
			} else {
				if loglv.IsDebug() {
					log.Printf("[DEBUG] killed the background contaienr: %s", cID)
				}
			}
		}()

		if loglv.IsDebug() {
			log.Printf("[DEBUG] booted the background contaienr: %s", cID)
		}

		// construct main command
		rawOptions := runtime.Options
		rawCommand := runtime.Command

		var options []string
		var command []string
		for _, v := range rawOptions {
			options = append(options, expand(v, user))
		}
		for _, v := range rawCommand {
			command = append(command, expand(v, user))
		}

		if sess.Command() != nil {
			command = sess.Command()
		}

		args = append(args, "exec", "-i")
		if isPty {
			args = append(args, "-t")
		}

		args = append(args, options...)
		args = append(args, cID)
		args = append(args, command...)
	} else {
		// just run a docker container

		// construct main command
		rawImage := runtime.Image
		rawOptions := runtime.Options
		rawCommand := runtime.Command

		// expand variables
		image := expand(rawImage, user)
		var options []string
		var command []string
		for _, v := range rawOptions {
			options = append(options, expand(v, user))
		}
		for _, v := range rawCommand {
			command = append(command, expand(v, user))
		}

		if sess.Command() != nil {
			command = sess.Command()
		}

		args = append(args, "run", "--rm", "-i", "-l", config.ContainerLabel)
		if isPty {
			args = append(args, "-t")
		}

		args = append(args, options...)
		args = append(args, image)
		args = append(args, command...)
	}

	if loglv.IsDebug() {
		log.Printf("[DEBUG] docker command args: %s", args)
	}

	cmd := exec.Command("docker", args...)

	if isPty {
		if loglv.IsDebug() {
			log.Printf("[DEBUG] running a container with allocated pty")
		}

		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		f, err := pty.Start(cmd)
		if err != nil {
			return exitStatus(err), err
		}

		go func() {
			for win := range winCh {
				setWinsize(f, win.Width, win.Height)
			}
		}()
		go func() {
			io.Copy(f, sess) // stdin
		}()
		io.Copy(sess, f) // stdout

	} else {
		if loglv.IsDebug() {
			log.Printf("[DEBUG] running a container")
		}

		cmd.Stdout = sess
		cmd.Stderr = sess.Stderr()

		// get stdin as a pipe
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return 1, err
		}
		go func() {
			io.Copy(stdin, sess)
			stdin.Close()
		}()

		if err := cmd.Start(); err != nil {
			return 1, err
		}

		err = cmd.Wait()

		return exitStatus(err), err
	}

	return 0, nil
}

func exitStatus(err error) int {
	var exitStatus int
	if err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			} else {
				exitStatus = 1
				log.Print("[error] unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus.")
			}
		}
	} else {
		exitStatus = 0
	}

	return exitStatus
}

func (srv *Server) runtime(name string) *RuntimeConfig {

	c := srv.Config

	if runtime, ok := c.Runtimes[name]; ok {
		return runtime
	}

	if fallback, ok := c.Runtimes["_fallback"]; ok {
		return fallback
	}

	return nil
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func (srv *Server) Close() error {
	return nil
}

func expand(value string, runtimeName string) string {
	return os.Expand(value, func(s string) string {
		switch s {
		case "SSHDOCKER_SSH_USER":
			return runtimeName
		}
		return os.Getenv(s)
	})
}

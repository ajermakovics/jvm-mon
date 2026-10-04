package jvmmon

import (
	"errors"
	"fmt"
	"github.com/tokuhirom/go-hsperfdata/attach"
	hs "github.com/tokuhirom/go-hsperfdata/hsperfdata"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type JVM struct {
	Pid      string
	ProcName string
	User     string
	Version  string
	socket   *attach.Socket
}

// GetCurUser returns the effective user name. Falls back to $USER when the
// user database lookup fails (e.g. static binary without a passwd entry).
func GetCurUser() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	return os.Getenv("USER")
}

func newRepository(user string) (*hs.Repository, error) {
	if user == "" {
		return hs.New()
	} else {
		return hs.NewUser(user)
	}
}

func attachPid(pid string, pidUser string) {
	curUser := GetCurUser()

	if curUser != pidUser && curUser == "root" { // root on linux

		pidDir := fmt.Sprintf("/proc/%s/cwd", pid)
		if ok, _ := exists(pidDir); !ok {
			return
		}
		usr, err := user.Lookup(pidUser)
		if err != nil {
			log.Println("Cannot lookup user", pidUser, err)
			return
		}
		uid, err1 := strconv.Atoi(usr.Uid)
		gid, err2 := strconv.Atoi(usr.Gid)
		if err1 != nil || err2 != nil {
			log.Println("Invalid uid/gid for user", pidUser, usr.Uid, usr.Gid)
			return
		}

		log.Println("Attaching to JVM of user id: ", uid, gid)
		attachFile := fmt.Sprintf("/proc/%s/cwd/.attach_pid%s", pid, pid)
		f, err := os.Create(attachFile)
		if err != nil {
			log.Printf("Cannot create file %v %v", attachFile, err)
			return
		}
		logErr("chown error ", os.Chown(attachFile, uid, gid))
		logErr("Cannot close", f.Close())
	}
}

func logErr(msg string, err error) {
	if err != nil {
		log.Println(msg, err)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (j *JVM) Attach() error {
	if j.Attached() {
		return nil
	}
	pidNr, err := strconv.Atoi(j.Pid)
	if err != nil {
		return fmt.Errorf("invalid pid %q: %w", j.Pid, err)
	}
	socketFile, _ := attach.GetSocketFile(pidNr)
	log.Println("Socket file:", socketFile)

	attachPid(j.Pid, j.User)

	sock, err := attach.New(pidNr)
	if err != nil {
		log.Println("Attach error ", err, "pid", pidNr)
		return err
	}
	log.Println("Attached to JVM")
	j.socket = sock
	return nil
}

func (j *JVM) Detach() error {
	sock := j.socket
	j.socket = nil
	if sock == nil {
		return nil
	}
	return sock.Close()
}

func (j *JVM) Attached() bool {
	return j.socket != nil
}

func (j *JVM) Properties() (string, error) {
	if j.socket == nil {
		return "", errors.New("not attached")
	}
	err := j.socket.Execute("properties")
	if err != nil {
		log.Println("Properties error ", err)
		return "", err
	}
	return j.socket.ReadString()
}

func (j *JVM) LoadAgent(agentJar string, args string) error {
	if j.socket == nil {
		return errors.New("not attached")
	}
	absolute := "false"
	agent := agentJar + "=" + args
	err := j.socket.Execute("load", "instrument", absolute, agent)
	if err != nil {
		log.Println("LoadAgent Execute error ", err)
		return err
	}
	out, er := j.socket.ReadString()
	if er != nil {
		log.Println("LoadAgent out error ", er)
		return er
	}
	log.Println("LoadAgent out ", out)
	return nil
}

func (j *JVM) AttachAndLoadAgent(jar string, args string) error {
	log.Println("Attaching to Pid:", j.Pid, "jar:", jar, "args:", args)
	err := j.Attach()
	if err != nil {
		log.Println("Cannot attach ", err)
		return err
	}

	log.Println("Loading agent ", jar, " ", args)
	err = j.LoadAgent(jar, args)
	if err == nil {
		log.Println("Loaded agent")
		return nil
	}
	log.Println("Load agent error ", err)
	logErr("Detach error ", j.Detach())
	return err
}

func GetJvmPidsByUser() (*map[string]string, error) {
	var users = make(map[string]string)
	numbers := regexp.MustCompile("^[0-9]+$")

	// hsperfdata dirs live directly in the temp dir: <tmp>/hsperfdata_<user>/<pid>
	dirs, err := filepath.Glob(filepath.Join(os.TempDir(), "hsperfdata_*"))
	if err != nil {
		return &users, err
	}
	for _, dir := range dirs {
		user := strings.TrimPrefix(filepath.Base(dir), "hsperfdata_")
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Println("Cannot read", dir, err)
			continue
		}
		for _, e := range entries {
			if e.Type().IsRegular() && numbers.MatchString(e.Name()) {
				users[e.Name()] = user
			}
		}
	}

	return &users, nil
}

func GetJVMUsers() []string {
	var userPids = make(map[string]string)

	pids, err := GetJvmPidsByUser()
	if err == nil {
		for pid, user := range *pids {
			userPids[user] = pid
		}
	} else {
		log.Println("Error finding JVMs: ", err)
	}

	var users []string
	for user := range userPids {
		users = append(users, user)
	}
	return users
}

func GetJVMs() map[string]JVM {
	jvms := map[string]JVM{}

	users := GetJVMUsers()
	log.Println("JVM users", len(users))

	for _, usr := range users {
		log.Println("Found JVM user", usr)
		userJvms, err := GetUserJVMs(usr)

		if err == nil {
			for pid, jvm := range userJvms {
				jvms[pid] = jvm
			}
		}
	}

	return jvms
}

func GetUserJVMs(user string) (map[string]JVM, error) {
	jvms := map[string]JVM{}

	repo, err := newRepository(user)
	if err != nil {
		log.Println("Cannot initialize ", err)
		return jvms, err
	}
	files, err := repo.GetFiles()
	if err != nil {
		println("No running JVMs found for user: ", user)
		log.Println("No JVMs found for user ", user, err)
		return jvms, err
	}

	for _, f := range files {
		res, err := f.Read()

		var jvm JVM
		if err == nil {
			procName := res.GetProcName()
			splitted := strings.Split(procName, string(os.PathSeparator))
			procName = splitted[len(splitted)-1]
			props := res.GetMap()
			jvmVer, _ := props["java.property.java.vm.specification.version"].(string)

			jvm = JVM{f.GetPid(), procName, user, jvmVer, nil}
		} else {
			procName := getCmdline(f.GetPid())
			jvm = JVM{f.GetPid(), procName, user, "", nil}
		}
		jvms[jvm.Pid] = jvm
	}

	return jvms, nil
}

func getCmdline(pid string) string {
	if runtime.GOOS != "linux" {
		return ""
	}
	cmdlinePath := filepath.Join("/proc", pid, "cmdline")
	cmdline, err := os.ReadFile(cmdlinePath)
	if err != nil {
		log.Println("Cannot read ", cmdlinePath, err)
		return ""
	}
	return strings.ReplaceAll(string(cmdline), "\x00", " ")
}

package jvmmon

import (
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

func GetCurUser() string {
	var user string
	if runtime.GOOS == "windows" {
		user = os.Getenv("USERNAME")
	} else {
		user = os.Getenv("USER")
	}
	return user
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
		exists, _ := exists(pidDir)

		if exists {
			usr, _ := user.Lookup(pidUser)
			uid, _ := strconv.Atoi(usr.Uid)
			gid, _ := strconv.Atoi(usr.Gid)

			log.Println("Attaching to JVM of user id: ", uid, gid)
			attachFile := fmt.Sprintf("/proc/%s/cwd/.attach_pid%s", pid, pid)
			f, err := os.Create(attachFile)

			if err != nil {
				log.Println(fmt.Sprintf("Canot create file %v %v", attachFile, err))
			} else {
				err := os.Chown(attachFile, uid, gid)
				logErr("chown error ", err)

			}
			logErr("Cannot close", f.Close())
		}
	}
}

func logErr(msg string, err error) {
	if err != nil {
		log.Println(msg, err)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func (j *JVM) Attach() error {
	if j.Attached() {
		return nil
	}
	pidNr, _ := strconv.Atoi(j.Pid)
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
	return sock.Close()
}

func (j *JVM) Attached() bool {
	return j.socket != nil
}

func (j *JVM) Properties() (string, error) {
	err := j.socket.Execute("properties")
	if err != nil {
		log.Fatal("Properties error ", err)
		return "", err
	}
	return j.socket.ReadString()
}

func (j *JVM) LoadAgent(agentJar string, args string) error {
	absolute := "false"
	agent := agentJar + "=" + args
	err := j.socket.Execute("load", "instrument", absolute, agent)
	if err != nil {
		log.Println("LoadAgent Execute error ", err)
		return err
	}
	out, er := j.socket.ReadString()
	if er != nil {
		log.Println("LoadAgent out error ", err)
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
		return err
	} else {
		log.Println("Load agent error ", err)
	}
	return j.Detach()
}

func GetJvmPidsByUser() (*map[string]string, error) {
	var users = make(map[string]string)
	numbers := regexp.MustCompile("[0-9]+")

	err := filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "hsperfdata_") && info.Mode().IsRegular() {
			parts := strings.Split(path, string(os.PathSeparator))
			pidFile := parts[len(parts)-1]

			if numbers.MatchString(pidFile) {
				userDir := parts[len(parts)-2]
				user := strings.Split(userDir, "_")[1]
				users[pidFile] = user
			}
		}
		return nil
	})

	return &users, err
}

func GetJVMUsers() []string {
	var userPids = make(map[string]string)

	pids, err := GetJvmPidsByUser()
	if err == nil {
		for pid, user := range *pids {
			userPids[user] = pid
		}
	} else {
		log.Fatal("Error finding JVMs: ", err)
	}

	var users []string
	for user, _ := range userPids {
		users = append(users, user)
	}
	return users
}

func GetJVMs() map[string]JVM {
	jvms := map[string]JVM{}

	users := GetJVMUsers()

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
			jvmVer := props["java.property.java.vm.specification.version"].(string)

			//log.Println("Props for " + f.GetPid(), " - ", procName)
			//for k, v := range res.GetMap() {
			//	log.Println("Proc data: ", k, v)
			//}
			//log.Println("---")

			jvm = JVM{f.GetPid(), procName, user, jvmVer, nil}
		} else {
			jvm = JVM{f.GetPid(), "", user, "", nil}
		}
		jvms[jvm.Pid] = jvm
	}

	return jvms, nil
}

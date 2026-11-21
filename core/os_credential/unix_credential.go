//go:build unix

// Package credential provides functions to switch group and user for command execution.
package credential

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	osUser "os/user"
	"strconv"
	"syscall"

	"go.uber.org/zap"
)

func Validate(log *zap.Logger, usr string, grp string) error {
	cu, err := osUser.Current()
	if err != nil {
		return fmt.Errorf("cannot get current user error: %w", err)
	}
	if usr != "" && cu.Uid != "0" {
		return errors.New("cannot switch user of tasks without root privilege, if you need to use user in tasks run crontab-go as user root")
	}
	_, _, err = lookupUIDAndGID(usr, log)
	if err != nil {
		return fmt.Errorf("cannot get uid and gid of user `%s` error: %w", usr, err)
	}
	_, err = lookupGID(grp)
	if err != nil {
		return fmt.Errorf("cannot get gid of group `%s` error: %w", grp, err)
	}

	return nil
}

func SetUser(log *zap.Logger, proc *exec.Cmd, usr string, grp string) {
	if usr == "" {
		log.Debug("no username given, running as current user")
		return
	}

	uid, gid, err := lookupUIDAndGID(usr, log)
	if err != nil {
		log.Panic("cannot get uid and gid of user", zap.String("user", usr), zap.Error(err))
	}
	if grp != "" {
		gid, _ = lookupGID(grp)
	}

	setUID(proc, uid, gid)
}

func lookupGID(grp string) (gid uint32, err error) {
	if grp == "" {
		return 0, nil
	}
	g, err := osUser.LookupGroup(grp)
	if err != nil {
		return 0, fmt.Errorf("cannot find group `%s`: %w", grp, err)
	}
	gidU, err := strconv.ParseUint(g.Gid, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(gidU), nil
}

func lookupUIDAndGID(usr string, _ *zap.Logger) (uid uint32, gid uint32, err error) {
	if usr == "" {
		return 0, 0, nil
	}
	u, err := osUser.Lookup(usr)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get uid and gid of user `%s`: %w", usr, err)
	}
	uidU, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	gidU, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return 0, 0, err
	}
	return uint32(uidU), uint32(gidU), nil
}

func setUID(
	proc *exec.Cmd,
	uid uint32,
	gid uint32,
) {
	attrib := &syscall.SysProcAttr{}
	proc.SysProcAttr = attrib
	proc.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}
}

// SetCrontabUser applies the execution account and its supplementary groups.
func SetCrontabUser(proc *exec.Cmd, account *osUser.User, group string) error {
	if account.Uid == strconv.Itoa(os.Geteuid()) && group == "" {
		return nil
	}
	uid, err := strconv.ParseUint(account.Uid, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid crontab user ID: %w", err)
	}
	gid, err := strconv.ParseUint(account.Gid, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid crontab group ID: %w", err)
	}
	primary := uint32(gid)
	if group != "" {
		primary, err = lookupGID(group)
		if err != nil {
			return err
		}
	}
	groups, err := account.GroupIds()
	if err != nil {
		return fmt.Errorf("resolve groups for crontab user %q: %w", account.Username, err)
	}
	supplementary := make([]uint32, 0, len(groups))
	for _, groupID := range groups {
		id, err := strconv.ParseUint(groupID, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid crontab supplementary group ID: %w", err)
		}
		supplementary = append(supplementary, uint32(id))
	}
	setUID(proc, uint32(uid), primary)
	proc.SysProcAttr.Credential.Groups = supplementary
	return nil
}

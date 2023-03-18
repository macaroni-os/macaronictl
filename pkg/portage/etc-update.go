/*
Copyright © 2021-2022 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package portage

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/logrusorgru/aurora"
	"github.com/macaroni-os/macaronictl/pkg/logger"
	"github.com/macaroni-os/macaronictl/pkg/utils"

	"gopkg.in/yaml.v3"
)

type EtcUpdateOpts struct {
	Quiet        bool
	AutomergeAll bool
	Paths        []string
	MaskPaths    []string
}

type EtcUpdateTask struct {
	Rootdir  string
	Opts     *EtcUpdateOpts
	Conf     *EtcUpdateConf
	FilesMap map[string][]string

	AutomergeAll bool
	DiscardAll   bool
}

type EtcUpdateConf struct {
	// mode - true for text, false for menu (support incomplete)
	ModeText bool `yaml:"mode" json:"mode"`
	// Whether to clear the term prior to each display or not
	ClearTerm bool `yaml:"clear_term" json:"mode"`
	// Whether trivial/comment changes should be automerged
	EuAutomerge bool `yaml:"eu_automerge" json:"eu_automerge"`
	// arguments used whenever rm is called
	RmOpts string `yaml:"rm_opts" json:"rm_opts"`
	// arguments used whenever mv is called
	MvOpts string `yaml:"mv_opts" json:"mv_opts"`
	// arguments used whenever cp is called
	CpOpts string `yaml:"cp_opts" json:"cp_opts"`
	// set the pager for use with diff commands (this will
	// cause the PAGER environment variable to be ignored)
	Pager string `yaml:"pager,omitempty" json:"pager,omitempty"`
	// To enable when is used an editor as diff commands
	UsingEditor bool `yaml:"using_editor,omitempty" json:"using_editor,omitempty"`
	// Command to use for diff
	DiffCommand string `yaml:"diff_command" json:"diff_command"`
	// Command to use for merging files
	MergeCommand string `yaml:"merge_command" json:"merge_command"`
}

func NewEtcUpdateOpts() *EtcUpdateOpts {
	return &EtcUpdateOpts{
		Quiet: false,
		Paths: []string{},
	}
}

func NewEtcUpdateConf() *EtcUpdateConf {
	return &EtcUpdateConf{
		ModeText:     true,
		ClearTerm:    false,
		EuAutomerge:  true,
		RmOpts:       "-i",
		MvOpts:       "-1",
		CpOpts:       "-i",
		Pager:        "",
		UsingEditor:  false,
		DiffCommand:  "diff -uN %file1 %file2",
		MergeCommand: "sdiff -s -o %merged %orig %new",
	}
}

func (c *EtcUpdateConf) String() string {
	data, _ := yaml.Marshal(c)
	return string(data)
}

func NewEtcUpdateConfFromFile(f string) *EtcUpdateConf {
	log := logger.GetDefaultLogger()
	ans := NewEtcUpdateConf()

	cEnvs, err := ParseEnvFile(f)
	if err != nil {
		log.Warning(fmt.Sprintf(
			"Error on parsing file %s: %s. Skipped.",
			f, err.Error()))
	} else {

		// Check mode env
		if val, ok := cEnvs["mode"]; ok {
			if val == "0" {
				ans.ModeText = true
			} else {
				ans.ModeText = false
			}
		}

		// Check clear_term
		if val, ok := cEnvs["clear_term"]; ok {
			if val == "yes" {
				ans.ClearTerm = true
			} else {
				ans.ClearTerm = false
			}
		}

		// Check eu_automerge
		if val, ok := cEnvs["eu_automerge"]; ok {
			if val == "yes" {
				ans.EuAutomerge = true
			} else {
				ans.EuAutomerge = false
			}
		}

		// Check rm_opts
		if val, ok := cEnvs["rm_opts"]; ok {
			ans.RmOpts = val
		}

		// Check mv_opts
		if val, ok := cEnvs["mv_opts"]; ok {
			ans.MvOpts = val
		}

		// Check cp_opts
		if val, ok := cEnvs["cp_opts"]; ok {
			ans.CpOpts = val
		}

		// Check pager
		if val, ok := cEnvs["pager"]; ok {
			ans.Pager = val
		}

		// Check using_editor
		if val, ok := cEnvs["using_editor"]; ok {
			if val == "1" {
				ans.UsingEditor = true
			} else {
				ans.UsingEditor = false
			}
		}

		// Check diff_command
		if val, ok := cEnvs["diff_command"]; ok {
			ans.DiffCommand = val
		}

		// Check merge_command
		if val, ok := cEnvs["merge_command"]; ok {
			ans.MergeCommand = val
		}
	}

	return ans
}

func EtcUpdate(rootdir string, opts *EtcUpdateOpts) error {
	var conf *EtcUpdateConf = nil
	log := logger.GetDefaultLogger()
	etcUpdateConffile := filepath.Join(rootdir, "/etc/etc-update.conf")

	// Prepare temporary directory used to store diff files.
	workDir, err := os.MkdirTemp(os.Getenv("PORTAGE_TMPDIR"), "etc-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	if utils.Exists(etcUpdateConffile) {
		conf = NewEtcUpdateConfFromFile(etcUpdateConffile)
	} else {
		conf = NewEtcUpdateConf()
	}

	if !conf.ModeText {
		return errors.New("Menu mode not supported yet.")
	}

	if conf.Pager == "" {
		// Check PAGER env
		if os.Getenv("PAGER") != "" {
			conf.Pager = os.Getenv("PAGER")
		} else {
			conf.Pager = "cat"
		}
	}

	if os.Getenv("NONINTERACTIVE_MV") != "" {
		conf.MvOpts = strings.ReplaceAll(conf.MvOpts, " -i ", "")
	}

	log.Debug(fmt.Sprintf("\n%s\n", conf))

	if !conf.UsingEditor {
		// Sanity check of the diff_command
		diffTestFile1 := filepath.Join(workDir, ".diff-test-1")
		diffTestFile2 := filepath.Join(workDir, ".diff-test-2")

		testData := []byte("test")
		err := os.WriteFile(diffTestFile1, testData, 0666)
		if err != nil {
			return err
		}
		err = os.WriteFile(diffTestFile2, testData, 0666)
		if err != nil {
			return err
		}
		testData = nil

		err = diffCommand(conf, diffTestFile1, diffTestFile2, true)
		if err != nil {
			return fmt.Errorf("%s does not seem to work, aborting",
				conf.DiffCommand,
			)
		}

	}

	if len(opts.Paths) == 0 {
		// POST: No user custom paths defined. Retrieve the
		//       protect config directories.

		configProtect := []string{}
		configProtectMask := []string{}

		profileEnvFile := filepath.Join(rootdir, "/etc/profile.env")
		if utils.Exists(profileEnvFile) {
			profEnvs, err := ParseEnvFile(profileEnvFile)
			if err != nil {
				return err
			}

			if val, ok := profEnvs["CONFIG_PROTECT"]; ok {
				configProtect = append(configProtect,
					strings.Split(val, " ")...)
			}

			if val, ok := profEnvs["CONFIG_PROTECT_MASK"]; ok {
				configProtectMask = append(configProtectMask,
					strings.Split(val, " ")...)
			}
		}

		if !utils.KeyInList("/etc", &configProtect) {
			// NOTE: Normally, it seems that on profile.env
			//       is not present /etc that instead is available
			//       with `portageq envvar -v CONFIG_PROTECT`
			configProtect = append([]string{"/etc"},
				configProtect...)
		}

		// Add Macaroni OS masks if not availables
		macaroniOsMask := []string{
			"/etc/macaroni/release",
			"/etc/os-release",
		}

		for _, mom := range macaroniOsMask {
			if !utils.KeyInList(mom, &configProtectMask) {
				configProtectMask = append(configProtectMask,
					mom,
				)
			}
		}

		opts.Paths = configProtect
		opts.MaskPaths = append(opts.MaskPaths, configProtectMask...)
	}

	task := &EtcUpdateTask{
		Opts:         opts,
		Conf:         conf,
		FilesMap:     make(map[string][]string, 0),
		Rootdir:      rootdir,
		AutomergeAll: false,
		DiscardAll:   false,
	}

	err = scan(rootdir, task)
	if err != nil {
		return err
	}

	if len(task.FilesMap) == 0 {
		// Nothing to do
		log.Info("Nothing left to do; exiting. :)")
		return nil
	}

	// Pre-analysis to automerge masked files
	err = checkMasked(task)
	if err != nil {
		return err
	}

	err = processFiles(task)
	if err != nil {
		return err
	}

	log.Info("Nothing left to do; exiting. :)")

	return nil
}

func checkMasked(task *EtcUpdateTask) error {
	log := logger.GetDefaultLogger()

	file2Remove := []string{}

	for f, cfgs := range task.FilesMap {

		base := f
		// Get rootfs full path.
		if task.Rootdir != "/" {
			if strings.HasPrefix(f, task.Rootdir) {
				base = base[len(task.Rootdir):]
			}
		}

		log.Debug("Checking", base, "...", task.Rootdir)

		if utils.KeyInList(base, &task.Opts.MaskPaths) {
			log.Info(fmt.Sprintf(
				"Automerging file %s config protect masked.", base))

			// Merge all files
			for _, c := range cfgs {
				err := os.Remove(f)
				if err != nil {
					return err
				}

				err = os.Rename(c, f)
				if err != nil {
					return err
				}
			}

			file2Remove = append(file2Remove, f)
		}

	}

	if len(file2Remove) > 0 {
		for _, f := range file2Remove {
			delete(task.FilesMap, f)
		}
	}

	return nil
}

func processFiles(task *EtcUpdateTask) error {
	log := logger.GetDefaultLogger()
	completed := false
	ask := ""

	for !completed {

		// Create file list and print message to user
		filesList := []string{}
		idx := 1

		fmt.Println(
			`The following is the list of files which need updating, each
configuration file is followed by a list of possible replacement files.
`)

		for k := range task.FilesMap {
			filesList = append(filesList, k)
			fmt.Println(fmt.Sprintf(
				"[%3d] %s", aurora.Bold(idx), aurora.Bold(k)))
			idx++
		}

		fmt.Println(`
[ -1] to exit
[ -3] to auto merge all files
[ -7] to discard all updates
`)

		fmt.Print(`
Please select a file to edit by entering the corresponding number.
	(don't use -3 or -7 if you're ensure what to do): `)

		_, err := fmt.Scanln(&ask)
		if err != nil {
			return err
		}

		res, err := strconv.Atoi(ask)
		if err != nil {
			return err
		}

		switch res {
		case -1:
			completed = true
			break
		case -3:
			task.AutomergeAll = true

			for _, f := range filesList {
				err = processFile(f, task)
				if err != nil {
					return err
				}
			}

		case -5:
			log.Warning("Not supported value. Try again.")
		case -7:
			task.DiscardAll = true
			for _, f := range filesList {
				err = processFile(f, task)
				if err != nil {
					return err
				}
			}
		case -9:
			log.Warning("Not supported value. Try again.")
		default:
			if res > idx {
				log.Warning(fmt.Sprintf(
					"Value '%s is not valid. Try again.", res))
			} else {
				err = processFile(filesList[res-1], task)
				if err != nil {
					return err
				}
			}
		}

		if len(task.FilesMap) == 0 {
			completed = true
		}
	}

	return nil
}

func processFile(f string, task *EtcUpdateTask) error {
	cfgs, ok := task.FilesMap[f]
	if !ok {
		return fmt.Errorf("Unexpected state. File %s not found", f)
	}

	log := logger.GetDefaultLogger()
	completed := false
	ask := ""

	if task.AutomergeAll || task.DiscardAll {

		for _, cfg := range cfgs {
			err := doCfg(f, cfg, task)
			if err != nil {
				return err
			}
			task.DelFileConfig(f, cfg)
		}

		return nil

	} else {

		for !completed {
			fmt.Println(fmt.Sprintf(`
Below are the new config files for %s:`, aurora.Bold(f)))

			for idx, cfg := range cfgs {
				fmt.Println(fmt.Sprintf(
					"[%3d] %s", aurora.Bold(idx+1),
					aurora.Bold(filepath.Base(cfg))))
			}

			fmt.Print(`
Please select a file to process (-1 to exit this file): `)

			_, err := fmt.Scanln(&ask)
			if err != nil {
				return err
			}

			res, err := strconv.Atoi(ask)
			if err != nil {
				return err
			}

			switch res {
			case -1:
				completed = true
				break
			default:
				if res > len(cfgs) {
					log.Warning(fmt.Sprintf(
						"Value '%s is not valid. Try again.", res))
					break
				}

				err = doCfg(f, cfgs[res-1], task)
				if err != nil {
					return err
				}
				task.DelFileConfig(f, cfgs[res-1])

				cfgs, ok = task.FilesMap[f]
				if !ok {
					// No more config
					completed = true
				}
			}
		}
	}

	return nil
}

func (task *EtcUpdateTask) DelFileConfig(f, cfg string) {
	val, ok := task.FilesMap[f]
	if ok {
		newval := []string{}
		for _, c := range val {
			if c != cfg {
				newval = append(newval, c)
			}
		}

		if len(newval) == 0 {
			delete(task.FilesMap, f)
		} else {
			task.FilesMap[f] = newval
		}
	}

}

func doCfg(ofile, cfg string, task *EtcUpdateTask) error {
	log := logger.GetDefaultLogger()

	completed := false
	ask := ""
	fabs := filepath.Join(task.Rootdir, ofile)
	cfgabs := filepath.Join(task.Rootdir, cfg)

	if task.DiscardAll {

		// POST: Delete update, keeping original file
		err := os.Remove(cfgabs)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("File %s deleted. :check_mark:",
			filepath.Base(cfg)))

		return nil
	}

	// Retrieve original file options/mode
	info, err := os.Stat(ofile)
	if err != nil {
		return err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("Unexpected error retrieve file stats for %s", ofile)
	}

	// Prepare temporary directory used to store diff files.
	workDir, err := os.MkdirTemp(os.Getenv("PORTAGE_TMPDIR"), "etc-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	mergeFunc := func() error {
		// Create merge file
		mergeFile := filepath.Join(workDir,
			fmt.Sprintf(".%s-%s-merged", filepath.Base(ofile), filepath.Base(cfg)))
		// Add empty line
		fmt.Println("")

		if task.Conf.UsingEditor {

			oFile := filepath.Join(workDir,
				fmt.Sprintf(".%s-copy", filepath.Base(ofile)))

			err = utils.CopyFile(cfgabs, mergeFile)
			if err != nil {
				return err
			}

			err = utils.CopyFile(fabs, oFile)
			if err != nil {
				return err
			}

			// Show diff
			err = diffCommand(task.Conf, oFile, mergeFile, false)
		} else {
			err = mergeCommand(task.Conf, mergeFile, fabs, cfgabs, false)
		}
		if err != nil {
			return err
		}

		replaceResp := ""
		fmt.Print("Replace original with merged file? (yes|y|n|no): ")

		_, err := fmt.Scanln(&replaceResp)
		if err != nil {
			return err
		}

		if replaceResp == "yes" || replaceResp == "y" {

			err = os.Remove(fabs)
			if err != nil {
				return err
			}
			// Here, i can't use rename because rename tries
			// to create an hardlink that doesn't work for
			// different devices.
			err = os.CopyFile(mergeFile, fabs)
			if err != nil {
				return err
			}

			if err = os.Chown(fabs, int(stat.Uid), int(stat.Gid)); err != nil {
				return err
			}

			if err = os.Chmod(fabs, info.Mode()); err != nil {
				return err
			}

			err = os.Remove(cfgabs)
			if err != nil {
				return err
			}
			err = os.Remove(mergeFile)
			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf(
				"File %s replaced by merged file.", ofile))

		} else {
			log.Info(fmt.Sprintf(
				"Merge operation for file %s cancelled.", ofile))
			return nil
		}

		return nil
	}

	if task.AutomergeAll {

		autoMergeFunc := func() error {
			err = os.Remove(fabs)
			if err != nil {
				return err
			}
			err = os.Rename(cfgabs, fabs)
			if err != nil {
				return err
			}

			if err = os.Chown(fabs, int(stat.Uid), int(stat.Gid)); err != nil {
				return err
			}

			if err = os.Chmod(fabs, info.Mode()); err != nil {
				return err
			}

			log.Info(fmt.Sprintf(
				"File %s replaced by new file.", ofile))

			return nil
		}

		return autoMergeFunc()
	}

	for !completed {

		fmt.Print(fmt.Sprintf(`
File: %s
[  1] Replace original with update
[  2] Delete update, keeping original as is
[  3] Interactively merge original with update
[  4] Show differences

Please select from the menu above (-1 to ignore this update): `, cfg))

		_, err := fmt.Scanln(&ask)
		if err != nil {
			return err
		}

		res, err := strconv.Atoi(ask)
		if err != nil {
			return err
		}

		switch res {

		case 1:
			// POST: Replace original with update
			err = os.Remove(fabs)
			if err != nil {
				return err
			}
			err = os.Rename(cfgabs, fabs)
			if err != nil {
				return err
			}

			if err = os.Chown(fabs, int(stat.Uid), int(stat.Gid)); err != nil {
				return err
			}

			if err = os.Chmod(fabs, info.Mode()); err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Original file %s replaced. :check_mark:", ofile))
			completed = true

		case 2:
			// POST: Delete update, keeping original file
			err = os.Remove(cfgabs)
			if err != nil {
				return err
			}
			log.Info(fmt.Sprintf("File %s deleted. :check_mark:",
				filepath.Base(cfg)))
			completed = true

		case 3:
			err = mergeFunc()
			if err != nil {
				return err
			}
			completed = true

		case 4:
			// Using always diff
			tmpconf := NewEtcUpdateConf()
			tmpconf.Pager = task.Conf.Pager
			err = diffCommand(tmpconf, fabs, cfgabs, false)
			if err != nil {
				return err
			}

		case -1:
			return nil
		}
	}

	return nil
}

func scanDir(dir string, task *EtcUpdateTask) error {
	log := logger.GetDefaultLogger()

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			err = scanDir(filepath.Join(dir, f.Name()), task)
			if err != nil {
				return err
			}
		}

		if !strings.HasPrefix(f.Name(), "._cfg") {
			log.Debug("File", f.Name(), "skipped.")
			continue
		}
		// POST: file match the regex

		words := strings.Split(f.Name(), "_")
		if len(words) < 3 {
			log.Debug("File", f.Name(), "skipped.")
			continue
		}

		forig := filepath.Join(dir, strings.Join(words[2:], ""))
		// Check if the file exists - #1
		if !utils.Exists(forig) {
			log.Info(fmt.Sprintf(
				"File %s is an orphan. Removing it directly...", f.Name()))
			err = os.Remove(filepath.Join(dir, f.Name()))
			if err != nil {
				return err
			}
			continue
		}

		err = scanFile(forig, task)
		if err != nil {
			return err
		}
	}

	return nil
}

func scanFile(file string, task *EtcUpdateTask) error {
	log := logger.GetDefaultLogger()
	dir := filepath.Dir(file)
	base := filepath.Base(file)
	regexCfgs := regexp.MustCompile(
		fmt.Sprintf(`^._cfg[0-9]+_%s$`, base))

	sanitizedFile := file
	if task.Rootdir != "/" {
		sanitizedFile = sanitizedFile[len(task.Rootdir):]
	}

	log.Debug("Scan file", sanitizedFile)
	// Check if the file is already been processed
	if _, ok := task.FilesMap[file]; ok {
		return nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if !regexCfgs.MatchString(f.Name()) {
			log.Debug("File", f.Name(), "skipped.")
			continue
		}

		if val, ok := task.FilesMap[sanitizedFile]; ok {
			task.FilesMap[sanitizedFile] = append(val,
				filepath.Join(dir, f.Name()))
		} else {
			task.FilesMap[sanitizedFile] = []string{filepath.Join(dir, f.Name())}
		}
	}

	return nil
}

func scan(rootfs string, task *EtcUpdateTask) error {
	log := logger.GetDefaultLogger()

	if !task.Opts.Quiet {
		log.Info("Scanning Configuration files...")
	}

	for _, p := range task.Opts.Paths {

		abspath := filepath.Join(rootfs, p)

		// Check if the path exists
		if !utils.Exists(abspath) {
			log.Debug(fmt.Sprintf("Path %s doesn't exists. Ignoring.", abspath))
			continue
		}

		// Check if the path is a file
		isDir, err := utils.IsDir(abspath)
		if err != nil {
			return err
		}

		if isDir {
			err = scanDir(abspath, task)
			if err != nil {
				return err
			}
		} else {

			isLink, err := utils.IsLink(abspath)
			if err != nil {
				return err
			}

			if !isLink {
				err = scanFile(abspath, task)
				if err != nil {
					return err
				}
			} // else ignoring links and dangerous cycles
		}
	}

	return nil
}

func diffCommand(conf *EtcUpdateConf, file1, file2 string, quiet bool) error {
	log := logger.GetDefaultLogger()

	diffCommand := conf.DiffCommand
	// Replace %file1 with the file path
	diffCommand = strings.ReplaceAll(diffCommand, "%file1", file1)
	// Replace %file2 with the file path
	diffCommand = strings.ReplaceAll(diffCommand, "%file2", file2)

	var cmd *exec.Cmd = nil

	args := strings.Split(diffCommand, " ")

	if !quiet && !conf.UsingEditor {
		entrypoint := []string{"/bin/bash", "-c"}
		pipeargs := strings.Join(args, " ") + " | " + conf.Pager
		args = append(entrypoint, pipeargs)
	}

	log.Debug("Diff command: ", args)

	cmd = exec.Command(args[0], args[1:]...)
	if !quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("Error on start %s command: %s", args[0:1],
			err.Error())
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Error on wait %s command: %s", args[0:1],
			err.Error())
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf(
			"%s command exiting with %d", cmd.ProcessState.ExitCode())
	}

	return nil
}

func mergeCommand(conf *EtcUpdateConf,
	fileMerged, fileOld, fileNew string, quiet bool) error {
	var cmd *exec.Cmd

	log := logger.GetDefaultLogger()

	mergeCommand := conf.MergeCommand
	// Replace %merged with the file path where merge data
	mergeCommand = strings.ReplaceAll(mergeCommand, "%merged", fileMerged)
	// Replace %orig with the old file path
	mergeCommand = strings.ReplaceAll(mergeCommand, "%orig", fileOld)
	// Replace %new with the new file path
	mergeCommand = strings.ReplaceAll(mergeCommand, "%new", fileNew)

	args := strings.Split(mergeCommand, " ")

	log.Debug("Merge command: ", args)

	cmd = exec.Command(args[0], args[1:]...)
	if !quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("Error on start %s command: %s", args[0:1],
			err.Error())
	}

	if args[0] == "sdiff" {
		// NOTE: sdiff exist with 0 when the files are the same
		//       1 if different
		//       2 if trouble / error (or exit with q)

		_ = cmd.Wait()

		if cmd.ProcessState.ExitCode() == 2 {
			return fmt.Errorf(
				"%s command exiting with %d", cmd.ProcessState.ExitCode())
		}

	} else {
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("Error on wait %s command: %s", args[0:1],
				err.Error())
		}

		if cmd.ProcessState.ExitCode() != 0 {
			return fmt.Errorf(
				"%s command exiting with %d", cmd.ProcessState.ExitCode())
		}
	}

	return nil
}

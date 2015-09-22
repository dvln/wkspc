// Copyright © 2015 Erik Brady <brady@dvln.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package wkspc works with 'dvln' workspaces.  This typically means routines
// to query or update a workspaces dvln "meta-data" can be found here.
package wkspc

import (
	"os"
	"path/filepath"

	"github.com/dvln/codebase"
	"github.com/dvln/devline"
	"github.com/dvln/out"
	"github.com/dvln/pkg"
	"github.com/dvln/util/dir"
	"github.com/dvln/util/file"
	//"github.com/dvln/vcs"
	globs "github.com/dvln/viper"
)

// Reader is targeted at reading workspace information (no wkspc meta-data
// updates will occur during read operations although logging of info is
// possible during such operations, eg: debugging/tracing)
type Reader interface {
	RootDir(path ...string) (string, error)
	LogDir() (string, error)
	TmpDir() (string, error)
	Codebase() (codebase.Defn, error)
	CodebaseRev() (pkg.Revision, error)
	Devline() (devline.Revision, error)
	Pkg(pkgName string, pkgID int) (pkg.Defn, error)
	PkgRev(pkgName string, pkgID int) (pkg.Revision, error)
	PkgDevline(pkgRev pkg.Revision, pkgID int) (devline.Revision, error)
	//FIXME: it's likely if workspace ops come through this pkg that these
	//       can be internal functions that are used within the context of
	//       the operation and not needed in the interface
	//DBDir() (string, error)
	//DB() (string, error)
	//VCSDir() (string, error)
	//VCSDataDir() (string, error)
	//VCSDataRepo() (vcs.Repo, error)
}

// Writer is meant for writing workspace data
type Writer interface {
	SetRootDir(rootDir string) error
	SetCodebase(codebase.Defn) error
	UpdCodebase(codebase.Defn) error
	//CommitCodebase() error
	//PushCodebase() error
	SetDevline(devline.Revision) error
	CommitDevline() error
	//PushDevline() error
	GetPkgRev(pkg.Revision) error // "Get" here relates to '% dvln get ..'
	PullPkgRev(pkg.Revision) error
	//CommitPkgRev(pkg.Defn) error
	//PushPkgRev(pkg.Defn) error
	RmPkg(pkg.Defn) error
}

// Info contains, well, information about the workspace (eg: wkspc.Info)
type Info struct {
	RootDir string `json:"rootDir,omitempty"`
	MetaDir string `json:"metaDir,omitempty"`
	LogDir  string `json:"logDir,omitempty"`
	TmpDir  string `json:"tmpDir,omitempty"`
	VCSDir  string `json:"vcsDir,omitempty"`
	DBDir   string `json:"dbDir,omitempty"`
	DB      string `json:"db,omitempty"`
	//multipkg.Hierarchy
	//multipkg.Pkgs
}

// doPkgWkspcGlobsInit sets up default settings for any variables/opts used
// for the dvln wkspc pkg... "globals" so to speak.  These are currently
// stashed in the 'globs' (viper) package at the default level (lowest
// priority essentially) and can be overriden via config file, CLI
// flags, and codebase, package and devline level overrides (in time).
// - Any dvln specific package can use globs to store "globals" which are then
//   effectively visible across all dvln commands and non-generic packages
// - Generic packages (eg: lib/dvln/out, or: github.com/dvln/out) should *NOT*
//   use globs (viper) for config so as to remain extremely generic (and usable
//   by the global community)... for such generic packages (or 3rd party/vendor
//   packages dvln is using the dvln or dvln get cmd can use globs (vipers) to
//   get DVLN configuration or opts/etc... and then it can use API's for those
//   generic package (or package variables) to config that pkg for dvln's use
//   as one can see with dvln/cmd/dvln.go init'ing the 'out' pkg on startup.
//
// Note: for any new CLI focused option for the dvln meta-cmd please check out
//       setupDvlnCmdCLIArgs() in cmds/dvln.go and for subcommands see related
//       setup<Name>CmdCLIArgs() located within each cmds/<name>.go file, I'd
//       suggest searching on the string "NewCLIOpts" to find those locations.
func doPkgWkspcGlobsInit() {
	// Note: this is currently in sections related to the scope of how the
	//       variable can be set, feel free to set subsections within those
	//       sections if needed (eg: path variables, clitool name vars)...
	//       essentially any grouping you see fit at this point but try and
	//       at least get the top level Section right

	// Section: ConstGlobal variables to store data (default value only, no overrides)
	// - please add them alphabetically and don't reuse existing opts/vars
	globs.SetDefault("wkspcMetaDirName", ".dvln")
	globs.SetDesc("wkspcMetaDirName", "name of dir under wkspc root where dvln cfg lives", globs.InternalUse, globs.ConstGlobal)

	// Section: InternalGlobal variables to store data (default value, can be changed by this pkg)
	// - please add them alphabetically and don't reuse existing opts/vars
	globs.SetDefault("wkspcRootDir", "none")
	globs.SetDesc("wkspcRootDir", "the workspace root directory, if one exists", globs.InternalUse, globs.InternalGlobal)

	globs.SetDefault("wkspcMetaDir", "")
	globs.SetDesc("wkspcMetaDir", "the .dvln dir under the workspace root dir, empty if no root", globs.InternalUse, globs.InternalGlobal)

	globs.SetDefault("wkspcLogDir", "")
	globs.SetDesc("wkspcLogDir", "the .dvln/log dir under the workspace root dir, empty if no root", globs.InternalUse, globs.InternalGlobal)

	globs.SetDefault("wkspcTmpDir", "")
	globs.SetDesc("wkspcTmpDir", "the .dvln/tmp dir under the workspace root dir, empty if no root", globs.InternalUse, globs.InternalGlobal)

	globs.SetDefault("wkspcVCSDir", "")
	globs.SetDesc("wkspcVCSDir", "the .dvln/vcs dir under the workspace root dir, empty if no root", globs.InternalUse, globs.InternalGlobal)

	globs.SetDefault("wkspcVCSDataDir", "")
	globs.SetDesc("wkspcVCSDataDir", "the .dvln/vcs/wkspc dir under the workspace root dir, empty if no root", globs.InternalUse, globs.InternalGlobal)

	globs.SetDefault("wkspcDBDir", "")
	globs.SetDesc("wkspcDBDir", "the .dvln/db dir under the workspace root dir, empty if no root", globs.InternalUse, globs.InternalGlobal)

	globs.SetDefault("wkspcDB", "")
	globs.SetDesc("wkspcDB", "the .dvln/db/wkspc.db file under the workspace root dir, empty if no root", globs.InternalUse, globs.InternalGlobal)

	// Section: BasicGlobal variables to store data (env, config file, default)
	// - please add them alphabetically and don't reuse existing opts/vars

	// Section: CLIGlobal class options, vars that can come in from the CLI
	// - Don't put these here, see cmds/dlvn.go and the cmds/cmdglobs.go file

	// Section: <add more sections as needed>
}

// At startup time we'll initialize settings for the workspace, same may be
// "constant"-like, others may be overridable via the users config file and env
func init() {
	doPkgWkspcGlobsInit()
}

// RootDir will return the path to the top-most workspace root directory
// based on our current path unless a "starting directory" path is given
// (all other params are ignored).  It will return the path to the
// workspace root if it found one, otherwise "" (a non-nil error implies
// there was an unexpected problem... ie: not being able to find a workspace
// root directory is NOT an error condition).  Note: if you wish to bypass
// cached workspace root info then use RootDirFind() directly instead.
func RootDir(path ...string) (string, error) {
	// wkspcRootDir starts out as "none", if something else we've calc'd it..
	wkspcRootDir := globs.GetString("wkspcrootDir")
	if wkspcRootDir != "none" {
		return wkspcRootDir, nil
	}
	// otherwise find it (this will always try and find it and cache it)
	return RootDirFind(path...)
}

// RootDirFind gets the workspaces root dir (if there is one), note that this will
// not use any "cached" values and it will store the result in globs (viper)
// under the "wkspcRootDir" key (you can access that with any upper/lower case
// as viper is case insensitive).
func RootDirFind(path ...string) (string, error) {
	startDir := ""
	var err error
	if path == nil || path[0] == "" {
		startDir, err = os.Getwd()
		if err != nil {
			return "", out.WrapErr(err, "Unable to find the workspace root directory (get current working dir failed)", 4100)
		}
	} else {
		startDir = path[0]
	}
	rootDir, err := dir.FindDirInOrAbove(startDir, globs.GetString("wkspcMetaDirName"))
	if err == nil {
		// Cache the root dir in "globs" (viper) memory if no error
		globs.Set("wkspcRootDir", rootDir)
	}
	return rootDir, err
}

// prepIfNotThere will create the dir if it doesn't exist and store info about
// it in viper (if a name is given), returns the full dir name and any
// error that occurs.  Note that "item" can be "" to indicate no sub dir.
func prepIfNotThere(directory bool, parent, item, name string) (string, error) {
	var err error
	fullPath := ""
	if item != "" {
		fullPath = filepath.Join(parent, item)
	} else {
		fullPath = parent
	}
	if name != "" {
		globs.Set(name, fullPath)
	}
	if fullPath != "" {
		if directory {
			err = dir.CreateIfNotExists(fullPath)
		} else {
			err = file.CreateIfNotExists(fullPath)
		}
	}
	return fullPath, err
}

// SetRootDir will set the given dir as the workspace root directory, it will also
// set some "derived" settings in globs (viper) such as:
//   wkspcMetaDir:    $wsroot/.dvln
//   wkspcLogDir:     $wsroot/.dvln/log
//   wkspcTmpDir:     $wsroot/.dvln/tmp
//   wkspcVCSDir:     $wsroot/.dvln/vcs
//   wkspcVCSDataDir: $wsroot/.dvln/vcs/wkspc
//   wkspcDBDir:      $wsroot/.dvln/db
//   wkspcDB:         $wsroot/.dvln/db/wkspc.db
// As other dirs or files are added relative to the .dvln/ wkspc meta-data dir
// they can be added here to "bootstap" them.  Note that if
func SetRootDir(rootDir string) error {
	if rootDir == "" {
		// if there is no root dir found allow it to be set to "" and bail
		globs.Set("wkspcRootDir", rootDir)
		return nil
	}
	directory := true
	_, err := prepIfNotThere(directory, rootDir, "", "wkspcRootDir")
	if err != nil {
		return err
	}
	wkspcMetaDir, err := prepIfNotThere(directory, rootDir, globs.GetString("wkspcMetaDirName"), "wkspcMetaDir")
	if err != nil {
		return err
	}
	_, err = prepIfNotThere(directory, wkspcMetaDir, "log", "wkspcLogDir")
	if err != nil {
		return err
	}
	_, err = prepIfNotThere(directory, wkspcMetaDir, "tmp", "wkspcTmpDir")
	if err != nil {
		return err
	}
	wkspcVCSDir, err := prepIfNotThere(directory, wkspcMetaDir, "vcs", "wkspcVCSDir")
	if err != nil {
		return err
	}
	wkspcVCSDataDir, err := prepIfNotThere(directory, wkspcVCSDir, "wkspc", "wkspcVCSDataDir")
	if err != nil {
		return err
	}
	_, err = prepIfNotThere(!directory, wkspcVCSDataDir, "static.dvln", "wkspcStaticDvln")
	if err != nil {
		return err
	}
	wkspcDBDir, err := prepIfNotThere(directory, wkspcMetaDir, "db", "wkspcDBDir")
	if err != nil {
		return err
	}
	_, err = prepIfNotThere(!directory, wkspcDBDir, "wkspc.db", "wkspcDB")
	if err != nil {
		return err
	}
	return nil
}

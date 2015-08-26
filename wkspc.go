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

	"github.com/dvln/out"
	"github.com/dvln/util/dir"
	globs "github.com/dvln/viper"
)

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
	globs.SetDefault("dvlnwkspcmetadatadir", ".dvln")
	globs.SetDesc("dvlnwkspcmetadatadir", "where dvln workspace config info lives", globs.ExpertUser, globs.ConstGlobal)

	globs.SetDefault("wkspcroot", "")
	globs.SetDesc("wkspcroot", "the workspace root directory, if one exists", globs.ExpertUser, globs.ConstGlobal)

	// Section: BasicGlobal variables to store data (env, config file, default)
	// - please add them alphabetically and don't reuse existing opts/vars

	// Section: CLIGlobal class options, vars that can come in from the CLI
	// - Don't put these here, see cmds/dlvn.go and the cmds/cmdglobs.go file
	//

	// Section: <add more sections as needed>
}

// At startup time we'll initialize settings for the workspace, same may be
// "constant"-like, others may be overridable via the users config file and env
func init() {
	doPkgWkspcGlobsInit()
}

// Root will return the path to the top-most workspace root directory
// based on our current path unless a "starting directory" path is given
// (all other params are ignored).  It will return the path to the
// workspace root if it found one, otherwise "" (a non-nil error implies
// there was an unexpected problem... ie: not being able to find a workspace
// root directory is NOT an error condition)
func Root(path ...string) (string, error) {
	// if we've already gotten a workspace root, use it
	wkspcRoot := globs.GetString("wkspcRoot")
	if wkspcRoot != "" {
		return wkspcRoot, nil
	}
	// otherwise get it (this will always try and find it)
	return RootFind(path...)
}

// RootFind gets the workspaces root dir (if there is one), note that this will
// not use any "cached" values and it will store the result in globs (viper)
// under the "wkspcRoot" key (you can access that with any upper/lower case
// as viper is case insensitive).
func RootFind(path ...string) (string, error) {
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
	rootDir, err := dir.FindDirInOrAbove(startDir, globs.GetString("dvlnwkspcmetadatadir"))
	if err == nil {
		// keep in mind that this can be "" if not in a dvln workspace
		SetRoot(rootDir)
	}
	return rootDir, err
}

// SetRoot will set the given dir as the workspace root directory, it will also
// set some "derived" settings in globs (viper) such as:
//   wkspcMetaDir: $wsroot/.dvln
//   wkspcLogDir:  $wsroot/.dvln/log
//   wkspcTmpDir:  $wsroot/.dvln/tmp
//   wkspcDBDir:   $wsroot/.dvln/db
//   wkspcDB:      $wsroot/.dvln/db/bolt.db
// As other dirs or files are added relative to the .dvln/ wkspc meta-data dir
// they can be added here to "bootstap" them.  Note that if
func SetRoot(dir string) {
	globs.Set("wkspcRoot", dir)
	if dir != "" {
		wkspcMetaDir := filepath.Join(dir, globs.GetString("dvlnwkspcmetadatadir"))
		globs.Set("wkspcMetaDir", wkspcMetaDir)
		wkspcLogDir := filepath.Join(wkspcMetaDir, "log")
		globs.Set("wkspcLogDir", wkspcLogDir)
		wkspcTmpDir := filepath.Join(wkspcMetaDir, "tmp")
		globs.Set("wkspcTmpDir", wkspcTmpDir)
		wkspcDBDir := filepath.Join(wkspcMetaDir, "db")
		globs.Set("wkspcDBDir", wkspcDBDir)
		wkspcDB := filepath.Join(wkspcDBDir, "bolt.db")
		globs.Set("wkspcDB", wkspcDB)
	} else {
		globs.Set("wkspcMetaDir", "")
		globs.Set("wkspcLogDir", "")
		globs.Set("wkspcTmpDir", "")
		globs.Set("wkspcDBDir", "")
		globs.Set("wkspcDB", "")
	}
}

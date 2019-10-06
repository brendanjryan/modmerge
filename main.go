package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const (
	tmpDir = "_modmerge"
)

var (
	outFile string
)

func init() {
	cmd.Flags().StringVarP(&outFile, "outfile", "o", "go.mod.new", "The name of the output file.")
}

var cmd = &cobra.Command{
	Use:   "modmerge -o <file> [<files>...]",
	Short: "Merge multiple go.mod files into a single file",
	Long:  "Merge multiple go.mod files into a single file",

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one go.mod file")
		}

		// check that we can open all files
		for _, a := range args {
			_, err := os.Stat(a)
			if err != nil {
				return fmt.Errorf("unable to open file %s - %s", a, err)
			}
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("reading module files...")
		md, err := readModules(args)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("merging module files...")
		final := combineModules(md)

		log.Printf("writng final result to %s...", outFile)
		err = writeRes(args[0], final, outFile)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Successfully wrote merged modules to", outFile)
	},
}

func writeRes(base string, mods map[string]string, dest string) error {

	bb, err := ioutil.ReadFile(base)
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", tmpDir)
	if err != nil {
		return err
	}

	defer func() {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Println(err)
		}
	}()

	err = ioutil.WriteFile(filepath.Join(dir, "go.mod"), bb, 0666)
	if err != nil {
		return err
	}

	// get a list of all package names
	var pkgs []string
	for m := range mods {
		pkgs = append(pkgs, m)
	}
	sort.Strings(pkgs)

	editArgs := []string{"mod", "edit"}
	for _, mod := range pkgs {
		editArgs = append(editArgs, "-require="+mod+"@"+mods[mod])
	}

	c := exec.Command("go", editArgs...)
	c.Stderr = os.Stderr
	c.Dir = dir
	err = c.Run()
	if err != nil {
		return err
	}

	return copyFile(filepath.Join(dir, "go.mod"), dest)
}

func copyFile(in string, out string) error {
	fb, err := ioutil.ReadFile(in)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(out, fb, 0666)
	if err != nil {
		return err
	}

	return nil
}

func combineModules(deps []map[string]string) map[string]string {
	res := map[string]string{}

	for _, d := range deps {
		for pkg, version := range d {
			maxVersion, ok := res[pkg]
			if !ok || semver.Compare(version, maxVersion) > 0 {
				res[pkg] = version
			}
		}
	}

	return res
}

func readModules(modFiles []string) ([]map[string]string, error) {

	var deps []map[string]string
	for _, mf := range modFiles {
		log.Println("reading file:", mf)
		md, err := ioutil.ReadFile(mf)
		if err != nil {
			return nil, err
		}

		mv, err := moduleVersions(md)
		if err != nil {
			return nil, err
		}

		deps = append(deps, mv)
	}

	return deps, nil
}

type mod struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

func moduleVersions(modData []byte) (map[string]string, error) {
	dir, err := ioutil.TempDir("", tmpDir)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Println(err)
		}
	}()

	err = ioutil.WriteFile(filepath.Join(dir, "go.mod"), modData, 0666)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	c := exec.Command("go", "list", "-m", "-json", "all")
	c.Stdout = &out
	c.Stderr = os.Stderr
	c.Dir = dir
	err = c.Run()
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(&out)
	mods := map[string]string{}
	for {
		m := &mod{}
		err := dec.Decode(m)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if m.Version != "" {
			mods[m.Path] = m.Version
		}
	}

	return mods, nil
}

// usage:
// modmerge <files>...
func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

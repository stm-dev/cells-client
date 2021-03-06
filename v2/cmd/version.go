package cmd

import (
	"log"
	"os"
	"runtime"
	"text/template"
	"time"

	hashivers "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/pydio/cells-client/v2/common"
)

var cellsVersionTpl = `{{.PackageLabel}}
 Version: 	{{.Version}}
 Built: 	{{.BuildTime}}
 Git commit: 	{{.GitCommit}}
 OS/Arch: 	{{.OS}}/{{.Arch}}
 Go version: 	{{.GoVersion}}
`

type CellsClientVersion struct {
	PackageLabel string
	Version      string
	BuildTime    string
	GitCommit    string
	OS           string
	Arch         string
	GoVersion    string
}

var (
	format string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Cells Client version information",
	Long: `
The version command simply shows the version that is currently running.

It also provides various utility sub commands than comes handy when manipulating software files. 
`,
	Run: func(cm *cobra.Command, args []string) {
		var t time.Time
		if common.BuildStamp != "" {
			t, _ = time.Parse("2006-01-02T15:04:05", common.BuildStamp)
		} else {
			t = time.Now()
		}

		sV := "N/A"
		if v, e := hashivers.NewVersion(common.Version); e == nil {
			sV = v.String()
		}

		cv := &CellsClientVersion{
			PackageLabel: common.PackageLabel,
			Version:      sV,
			BuildTime:    t.Format(time.RFC822Z),
			GitCommit:    common.BuildRevision,
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			GoVersion:    runtime.Version(),
		}

		var runningTmpl string

		if format != "" {
			runningTmpl = format
		} else {
			// Default version template
			runningTmpl = cellsVersionTpl
		}

		tmpl, err := template.New("cells").Parse(runningTmpl)
		if err != nil {
			log.Fatalln("failed to parse template", err)
		}

		if err = tmpl.Execute(os.Stdout, cv); err != nil {
			log.Fatalln("could not execute template", err)
		}
	},
}

var ivCmd = &cobra.Command{
	Use:   "isvalid",
	Short: "Return an error if the passed version is not correctly formatted",
	Long: `Tries to parse the passed string version using the hashicorp/go-version library 
and hence validates that it respects semantic versionning rules.

In case the passed version is *not* valid, the process exits with status 1.`,
	Example: `
# A valid version
` + os.Args[0] + ` version isvalid 2.0.6-dev.20191205

# A *non* valid version
` + os.Args[0] + ` version isvalid 2.a
`,
	Run: func(cm *cobra.Command, args []string) {
		if len(args) != 1 {
			cm.Printf("Please provide a version to parse\n")
			os.Exit(1)
		}

		versionStr := args[0]

		_, err := hashivers.NewVersion(versionStr)
		if err != nil {
			cm.Printf("[%s] is *not* a valid version\n", versionStr)
			os.Exit(1)
			// do not output anything is case the version is correct.
			// } else {
			// 	cm.Printf("[%s] is a valid version\n", versionStr)
		}
	},
}

var irCmd = &cobra.Command{
	Use:   "isrelease",
	Short: "Return an error if the passed version is a snapshot",
	Long: `Tries to parse the passed string version using the hashicorp/go-version library 
and hence validates that it respects semantic versionning rules.

It then insures that the passed string is not a pre-release, 
that is that is not suffixed by "a hyphen and a series of dot separated identifiers 
immediately following the patch version", see: https://semver.org

In case the passed version is *not* a valid realease version, the process exits with status 1.`,
	Example: `
# A valid release version
` + os.Args[0] + ` version isvalid 2.0.6

# A *non* release version
` + os.Args[0] + ` version isvalid 2.0.6-dev.20191205
`,
	Run: func(cm *cobra.Command, args []string) {
		if len(args) != 1 {
			cm.Printf("Please provide a single version to be parsed\n")
			os.Exit(1)
		}

		versionStr := args[0]

		v, err := hashivers.NewVersion(versionStr)
		if err != nil {
			cm.Printf("[%s] is *not* a valid version\n", versionStr)
			os.Exit(1)
		}

		if v.Prerelease() != "" {
			// This is a pre-release, throw an error
			cm.Printf("[%s] is *not* a valid release version\n", versionStr)
			os.Exit(1)
		}
	},
}

var igtCmd = &cobra.Command{
	Use:   "isgreater",
	Short: "Compares the two passed versions and returns an error if the first is *not* strictly greater than the second",
	Long: `Tries to parse the passed string versions using the hashicorp/go-version library and returns an error if:
  - one of the 2 strings is not a valid semantic version,
  - the first version is not strictly greater than the second`,
	Example: `
# This exits with status 1.
` + os.Args[0] + ` version isgreater 2.0.6-dev.20191205 2.0.6
`,
	Run: func(cm *cobra.Command, args []string) {
		if len(args) != 2 {
			cm.Printf("Please provide two versions to be compared\n")
			os.Exit(1)
		}

		v1Str := args[0]
		v2Str := args[1]
		// fmt.Printf("Comparing versions %s & %s \n", v1Str, v2Str)

		v1, err := hashivers.NewVersion(v1Str)
		if err != nil {
			cm.Printf("Passed version [%s] is not a valid version\n", v1Str)
			os.Exit(1)
		}
		v2, err := hashivers.NewVersion(v2Str)
		if err != nil {
			cm.Printf("Passed version [%s] is not a valid version\n", v2Str)
			os.Exit(1)
		}
		if !v1.GreaterThan(v2) {
			cm.Printf("Passed version [%s] is *not* greater than [%s]\n", v1Str, v2Str)
			os.Exit(1)
		}
	},
}

func init() {
	versionCmd.AddCommand(ivCmd)
	versionCmd.AddCommand(irCmd)
	versionCmd.AddCommand(igtCmd)
	RootCmd.AddCommand(versionCmd)

	versionCmd.Flags().StringVarP(&format, "format", "f", "", "Use go template to format version output")
}

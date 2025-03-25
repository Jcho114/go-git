package cmd

import (
	"fmt"
	"os/user"
	"strconv"
	"time"

	"github.com/Jcho114/go-git/index"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

func init() {
	lsFilesCmd.Flags().BoolVar(&verbose, "verbose", false, "actually write object to database")
	rootCmd.AddCommand(lsFilesCmd)
}

var lsFilesCmd = &cobra.Command{
	Use:   "ls-files",
	Short: "a very attempt at displaying the files in the staging area",
	Long:  "a very very bad attempt at displaying the files in the staging area from scratch",
	Args:  cobra.NoArgs,
	RunE:  runLsFiles,
}

func runLsFiles(cmd *cobra.Command, args []string) error {
	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}
	ind, err := index.IndexRead(repository)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("index file format v%d, containing %d entries\n", ind.Version, len(ind.Entries))
	}

	for _, entry := range ind.Entries {
		fmt.Printf("%s\n", entry.Name)
		if verbose {
			var entrytype string
			switch entry.Modetype {
			case 0b1000:
				entrytype = "regular file"
			case 0b1010:
				entrytype = "symbolic link"
			case 0b1110:
				entrytype = "git link"
			}

			fmt.Printf("  %s with perms: %o\n", entrytype, entry.Modeperms)
			fmt.Printf("  on blob: %s\n", entry.Sha)
			ctime := time.Unix(entry.Ctime.Seconds, entry.Ctime.Nanoseconds)
			mtime := time.Unix(entry.Mtime.Seconds, entry.Mtime.Nanoseconds)
			dformat := "2018-12-25 09:27:53.000000000"
			fmt.Printf("  created: %s, modified: %s\n", ctime.Format(dformat), mtime.Format(dformat))
			fmt.Printf("  device: %d, inode: %d\n", entry.Dev, entry.Ino)
			us, err := user.LookupId(strconv.Itoa(entry.Uid))
			if err != nil {
				return err
			}
			gr, err := user.LookupGroupId(strconv.Itoa(entry.Gid))
			if err != nil {
				return err
			}
			fmt.Printf("  user: %s (%d)  group: %s (%d)\n", us.Name, entry.Uid, gr.Name, entry.Gid)
			fmt.Printf("  flags: stage=%d assume_valid=%t\n", entry.Flagstage, entry.Flagvalid)
		}
	}

	return nil
}

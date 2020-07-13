package command

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/format"
	"github.com/simplesurance/baur/format/csv"
	"github.com/simplesurance/baur/format/table"
	"github.com/simplesurance/baur/internal/command/term"
)

func init() {
	lsCmd.AddCommand(&newLsInputsCmd().Command)
}

type lsInputsCmd struct {
	cobra.Command

	csv        bool
	quiet      bool
	showDigest bool
}

func newLsInputsCmd() *lsInputsCmd {
	cmd := lsInputsCmd{
		Command: cobra.Command{
			Use:   "inputs <APP-NAME>.<TASK-NAME>]",
			Short: "list resolved task inputs of an application",
			Args:  cobra.ExactArgs(1),
		},
	}

	cmd.Run = cmd.run

	cmd.Flags().BoolVar(&cmd.csv, "csv", false,
		"Show output in RFC4180 CSV format")

	cmd.Flags().BoolVarP(&cmd.quiet, "quiet", "q", false,
		"Only show filepaths")

	cmd.Flags().BoolVar(&cmd.showDigest, "digests", false,
		"show digests")

	return &cmd
}

func (c *lsInputsCmd) run(cmd *cobra.Command, args []string) {
	var formatter format.Formatter
	var headers []string

	rep := mustFindRepository()
	task := mustArgToTask(rep, args[0])
	writeHeaders := !c.quiet && !c.csv

	if !task.HasInputs() {
		stderr.TaskPrintf(task, "has no inputs configured")
		exitFunc(1)
	}

	if writeHeaders {
		headers = []string{"Path"}

		if c.showDigest {
			headers = append(headers, "Digest")
		}
	}

	if c.csv {
		formatter = csv.New(headers, stdout)
	} else {
		formatter = table.New(headers, stdout)
	}

	inputResolver := baur.NewInputResolver()

	inputs, err := inputResolver.Resolve(rep.Path, task)
	exitOnErr(err)

	sort.Slice(inputs.Files, func(i, j int) bool {
		return inputs.Files[i].RepoRelPath() < inputs.Files[j].RepoRelPath()
	})

	for _, input := range inputs.Files {
		if !c.showDigest || c.quiet {
			mustWriteRow(formatter, input)
			continue
		}

		digest, err := input.Digest()
		exitOnErrf(err, "%s: calculating digest failed", input)

		mustWriteRow(formatter, input, digest.String())
	}

	err = formatter.Flush()
	exitOnErr(err)

	if c.showDigest && !c.quiet && !c.csv {
		totalDigest, err := inputs.Digest()
		exitOnErr(err, "calculating total input digest failed")

		stdout.Printf("\nTotal Input Digest: %s\n", term.Highlight(totalDigest.String()))
	}
}

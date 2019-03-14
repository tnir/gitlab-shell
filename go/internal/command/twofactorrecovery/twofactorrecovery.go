package twofactorrecovery

import (
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/reporting"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/twofactorrecovery/twofactorrecoveryclient"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
)

type Command struct {
	Config *config.Config
	Args   *commandargs.CommandArgs
}

func (c *Command) Execute(reporter *reporting.Reporter) error {
	if c.canContinue(reporter) {
		c.displayRecoveryCodes(reporter)
	} else {
		fmt.Fprintln(reporter.Out, "\nNew recovery codes have *not* been generated. Existing codes will remain valid.")
	}

	return nil
}

func (c *Command) canContinue(reporter *reporting.Reporter) bool {
	fmt.Fprintln(reporter.Out, "Are you sure you want to generate new two-factor recovery codes?")
	fmt.Fprintln(reporter.Out, "Any existing recovery codes you saved will be invalidated. (yes/no)")

	var answer string
	fmt.Fscanln(reporter.In, &answer)

	return (answer == "yes")
}

func (c *Command) displayRecoveryCodes(reporter *reporting.Reporter) {
	codes, err := c.getRecoveryCodes()

	if err == nil {
		fmt.Fprint(reporter.Out, "\nYour two-factor authentication recovery codes are:\n\n")
		fmt.Fprintln(reporter.Out, strings.Join(codes, "\n"))
		fmt.Fprintln(reporter.Out, "\nDuring sign in, use one of the codes above when prompted for")
		fmt.Fprintln(reporter.Out, "your two-factor code. Then, visit your Profile Settings and add")
		fmt.Fprintln(reporter.Out, "a new device so you do not lose access to your account again.")
	} else {
		fmt.Fprintf(reporter.Out, "\nAn error occurred while trying to generate new recovery codes.\n%v\n", err)
	}
}

func (c *Command) getRecoveryCodes() ([]string, error) {
	client, err := twofactorrecoveryclient.GetClient(c.Config)

	if err != nil {
		return nil, err
	}

	return client.GetRecoveryCodes(c.Args.GitlabKeyId)
}

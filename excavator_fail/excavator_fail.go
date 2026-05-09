package fail

fail

/*
This is a non-compiling file that has been added to explicitly ensure that CI fails.
It also contains the command that caused the failure and its output.
Remove this file if debugging locally.

./godelw verify failed after updating godel plugins and assets

Command that caused error:
./godelw verify --skip-test --skip-lint --skip-conjure-backcompat

Output:
Running format...
Running mod...
Running conjure...
Error: Get "https://repo1.maven.org/maven2/com/palantir/witchcraft/api/witchcraft-health-api/1.3.0/witchcraft-health-api-1.3.0.conjure.json": dial tcp 104.18.18.12:443: i/o timeout
Running license...
Running distgo-task...
Failed tasks:
	conjure

*/

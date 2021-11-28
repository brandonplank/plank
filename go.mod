module brandonplank.org/plank

go 1.17

require (
	github.com/akamensky/argparse v1.3.1
	brandonplank.org/plankcore v0.0.0
)

replace (
	brandonplank.org/plankcore => ./core
)

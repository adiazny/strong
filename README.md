# strong
strong app workout logs published to strava

# Program Behavior
- Get the strong app csv workout log file from Google Drive API
- Read the strong.csv workout log file
- Parse and convert workouts to Go workout structs
- Send workouts to Strava API

## The Power of Go Tests

"write calling code that is sensible, readable, and useful, and only then proceed to implement everything necessary to make it work."

"The most important reason to write tests first is that, to do that, we need to have a clear idea of how the program should behave, from the user’s point of view. There’s some thinking involved in that, and the best time to do it is before we’ve written any code."


# Google Drive API
- [Google Drive API](https://developers.google.com/drive/api) - [Google Workspace](https://developers.google.com/workspace/guides/get-started) - [Google Drive Activity API](https://developers.google.com/drive/activity/v2)

## Research
- VSCode extention to work with csv files: Edit CSV
- Golang csv libraries
    - Standard library: https://pkg.go.dev/encoding/csv@go1.19.3
easy-working-with-csv-in-golang-using-gocsv-package-9c8424728bbe
    - External:
        - https://github.com/gocarina/gocsv
                - https://articles.wesionary.team/read-and-write-csv-file-in-go-b445e34968e9
                - https://articles.wesionary.team/
        - https://github.com/jszwec/csvutil
# PSGradeUtility
a simple PowerSchool grade calculation utility, to optimize your laziness

* built using PowerSchool's (undocumented) native mobile api, no html parsing!

## using
PSGradeUtility is built in 100% go and only has 1 external dependency, so it'll most likely work on almost any OS you need! i'll provide binaries for windows/macos/linux amd64/arm64 in releases

PSGradeUtility also defaults to using my school district's PowerSchool instance. if you need to change this, just edit `PowerSchoolInstance` in `main.go`.

1. download the binary for your OS from [releases](https://github.com/unmistakenly/PSGradeUtility/releases/latest) and launch
2. you can use `h` in any part of the program for help. you'll need to sign in first using `s` before fetching your grades with `a` or entering the grade calculator with `c`
3. after entering the grade calculator, just enter the number associated with the class you're looking for
4. see `h` and read for yourself how to add/delete your own grades

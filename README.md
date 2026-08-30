# PSGradeUtility
a simple PowerSchool grade calculation utility, to optimize your laziness

* built using PowerSchool's (undocumented) native mobile api, no html parsing!

## using
PSGradeUtility is built in 100% go and only has 1 external dependency, so it'll most likely work on almost any OS you need! i'll provide binaries for windows/macos/linux amd64/arm64 in releases

you need to point it at your own district's PowerSchool instance — nothing is baked in:

```
psgradeutility -instance https://myps.<your-district>.org
```

or set `PSGRADE_INSTANCE` in your environment instead of passing the flag every time.

1. download the binary for your OS from [releases](https://github.com/unmistakenly/PSGradeUtility/releases/latest) and launch with `-instance` set (or `PSGRADE_INSTANCE`)
2. you can use `h` in any part of the program for help. you'll need to sign in first using `s` before fetching your grades with `a` or entering the grade calculator with `c`
3. after entering the grade calculator, just enter the number associated with the class you're looking for
4. see `h` and read for yourself how to add/delete your own grades

## stupid notice(s)
i completely forgot to mention that this only works for districts using weighted grading systems. like 20% low, 30% mid, and 50% high. i forgot that other systems might exist — if yours weights differently, override it: `-low-weight -mid-weight -high-weight` (defaults are 0.2/0.3/0.5). the PowerSchool response doesn't appear to expose the actual weight percentages anywhere we could find, only category names — if you spot one while testing, open an issue, that'd let this derive it automatically instead of needing the flags.

## manual verification (no CI creds — do this once against a real account before relying on a release)
CI covers everything that doesn't need a live PowerSchool session (grade math, digest crypto, XML/response parsing). The actual sign-in + fetch flow needs a real account to exercise end to end:
1. `-instance` (or `PSGRADE_INSTANCE`) pointed at your district, launch, `s` to sign in with a real student account.
2. `a` — confirm it fetches and prints real grades without error.
3. `c` — enter the calculator, `add`/`edit`/`restore`/`del` a grade, confirm the final % updates as expected and `restore` reverts exactly.
4. `q` from inside the calculator — confirm it exits cleanly (no half-finished output, exit code 0).
5. Sign out (`o`) and sign back in with a parent account if you have one — confirm it's rejected (parent accounts are unsupported).

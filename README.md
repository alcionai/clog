# The Clues Logger!

## Regular logging
```go
clog.Ctx(ctx).Info("information")
clog.Ctx(ctx).Debug("debugging")
clog.Ctx(ctx).Err(err).Error("badness")
```
## Labeling your logs
```go
clog.CtxErr(ctx, err).
  Label(clog.LStartOfRun, clog.LFailureSource).
  Info("couldn't start up process")
```
## Commenting your logs 
```go
clog.Ctx(ctx).
  Comment(`If I wanted, i could add this all to code; but now we can pull double duty:
  first - whatever I say here is readable to anyone who is looking at the logs (which is good
  if i'm trying to tell them what they need to know about due to this log occurring)
  second - it's also a regular comment, as if in code, so the code is now also commented!`).
  Info("important things")
```
## Adding structured data
```go
ctx := clues.Add(ctx, "foo", "bar")
err := clues.New("a bad happened").With("fnords", "smarf")

clog.CtxErr(ctx, err).
  With("beaux", "regarde").
  Debug("all the info!")
// will output a log containing: 
// {
//  "msg": "all the info!",
//  "foo": "bar",
//  "fnords": "smarf",
//  "beaux": "regarde",
// }
```
## Setting up logs
```go
set := clog.Settings{
  Format: clog.LFHuman,
  Level: clog.LLInfo,
}

ctx := clog.Init(ctx, set)
```
## Filtering Debug Logs (aka, improved debug levels)
```go
set := clog.Settings{
  Format: clog.LFHuman,
  Level: clog.LLDebug,
  OnlyLogDebugIfContainsLabel: []string{clog.LAPICall},
}

ctx := clog.Init(ctx, set)
```

gunk
====

## DEPRECATION NOTICE

This repo has been deprecated/split apart. Its components have been
moved to the following destinations as submodules:

* gunk/http_wrap => [volman](https://github.com/cloudfoundry-incubator/volman)
* gunk/os_wrap   => [volman](https://github.com/cloudfoundry-incubator/volman)

* gunk/diegonats => [route-emitter](https://github.com/cloudfoundry/route-emitter)


Some components have also been turned into new repos:

* gunk/command_runner => [cloudfoundry/commandrunner](https://github.com/cloudfoundry/commandrunner)
* gunk/urljoiner      => [cloudfoundry/urljoiner](https://github.com/cloudfoundry/urljoiner)
* gunk/workpool       => [cloudfoundry/workpool](https://github.com/cloudfoundry/workpool)




The following repos are no longer used anywhere, and so have not been spun out:

* gunk/group_runner
* gunk/test_server
* gunk/natsrunner
* gunk/runner_support
* gunk/natbeat
* gunk/metricz        => [cloudfoundry-attic/metricz](https://github.com/cloudfoundry-attic/metricz)




## Description


A collection of go fakes and their real counterparts

Recently timeprovider was removed from this repo. Use pivotal-golang/clock instead.

Detailed docs here:
[godoc](https://godoc.org/github.com/cloudfoundry/gunk)

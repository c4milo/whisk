# Whisk

Detects and finds circular dependencies in Chef's cookbooks.

- [Why](#why)
- [How](#how)
- [Usage](#usage)
- [Deployments](#deployments)
  - [Linux](#linux)
  - [MacOS](#macos)
- [Integrations](#integrations)

## Why

We built this tool to help introduce Chef [Policyfiles](https://docs.chef.io/policyfile/) into our chef-repo, which requires our dependency graph to be a DAG.

## How

A presentation about how this tool was built can be seen [here](https://docs.google.com/presentation/d/113TqRIfvJ6RpIpnwhcFTvZLRmOwS3IIeIL8ztJZ134g/edit#slide=id.g62bc63a727_0_0)

## Usage

```
$ ./whisk --help

Usage:
  whisk [flags] <role_path>
  whisk [command]

Available Commands:
  help        Help about any command
  lint        Lints all Chef roles dependencies to make sure a minimum quality bar is held

Flags:
  -c, --cookbook-path string   Comma-separated cookbook paths (default "./cookbooks")
  -h, --help                   help for whisk
  -o, --output string          Output format, either ascii, json or dot (default "ascii")

Use "whisk [command] --help" for more information about a command.
```

Example:

```
 $ ./whisk --cookbook-path=/Users/caguilar/src/chef-repo/site-cookbooks/,/Users/caguilar/src/chef-repo/cookbooks/ /Users/caguilar/src/chef-repo/roles/slack-min.json
.
└── slack-min
    ├── apt-freight
    │   ├── apt
    │   └── feature
    ├── awscli
    │   └── poise-python
    ├── chef-client
    │   ├── bluepill
    │   │   └── rsyslog
    │   └── slack-base-runit
    │       └── slack-runit
    │           └── runit
    ├── chef-handlers
    │   ├── chef_handler
    │   └── slack-base-helper
    ├── consul
    │   ├── aws-ro
    │   │   └── consul-template
    │   │       ├── nebula
    │   │       │   ├── slack-base-check
    │   │       │   ├── slack-collectd
    │   │       │   ├── slack-deployable
    │   │       │   │   ├── prometheus_targets
    │   │       │   │   │   ├── slack-cron
    │   │       │   │   │   ├── slack-instance
    │   │       │   │   │   ├── slackops-tool
    │   │       │   │   │   │   ├── s3_file
    │   │       │   │   │   │   ├── slack-base-update-template
    │   │       │   │   │   │   └── slack-python
    │   │       │   │   │   └── systemd-service
    │   │       │   │   ├── sudo
    │   │       │   │   └── what-happened
    │   │       │   │       └── slack-base-cron
    │   │       │   └── slack-go
    │   │       └── slack-validation
    │   ├── certs
    │   ├── credentials
    │   ├── slack-base-log
    │   │   ├── murron
    │   │   ├── slack-monitoring
    │   │   │   ├── perl
    │   │   │   │   └── windows
    │   │   │   └── slack-apache
    │   │   │       ├── apache2
    │   │   │       │   └── php
    │   │   │       ├── misc_env
    │   │   │       │   └── swap
    │   │   │       └── slack-git
    │   │   └── slack-monitoring-client
    │   └── slack-chef-client
    │       └── blocklist-node-attrs
    ├── discover
    ├── foodcourt
    ├── go-audit
    ├── hosts
    ├── load-balancer
    ├── ntp
    ├── slack-base-image
    │   └── ssh-known-hosts
    ├── slack-chef-metrics
    ├── slack-environment-guard
    ├── slack-logrotate
    ├── slack-motd
    ├── slack-nessus-agent
    ├── slack-rsyslog
    ├── slack-sysctl
    │   └── sysctl
    ├── slack-wrappers
    ├── sshd
    ├── ssm-agent
    ├── trace
    ├── users
    └── zsh


⚠️  Strongly Connected Components (topologically sorted): 1

1. what-happened, slack-monitoring-client, slack-apache, slackops-tool, prometheus_targets, slack-monitoring, murron, slack-base-log, certs, consul, slack-deployable, slack-collectd, nebula, consul-template, aws-ro, php, apache2


🌀 Cycles: 53

1. apache2, php, aws-ro, consul-template, nebula, slack-deployable, consul, slack-base-log, slack-monitoring, slack-apache, apache2
2. apache2, php, aws-ro, consul-template, nebula, slack-deployable, prometheus_targets, consul, slack-base-log, slack-monitoring, slack-apache, apache2
3. aws-ro, consul-template, nebula, slack-deployable, consul, aws-ro
4. aws-ro, consul-template, nebula, slack-deployable, consul, slack-base-log, murron, aws-ro
5. aws-ro, consul-template, nebula, slack-deployable, consul, slack-base-log, slack-monitoring, prometheus_targets, aws-ro
6. aws-ro, consul-template, nebula, slack-deployable, consul, slack-base-log, slack-monitoring, slack-apache, php, aws-ro
7. aws-ro, consul-template, nebula, slack-deployable, consul, slack-base-log, slack-monitoring-client, prometheus_targets, aws-ro
8. aws-ro, consul-template, nebula, slack-deployable, prometheus_targets, aws-ro
9. aws-ro, consul-template, nebula, slack-deployable, prometheus_targets, consul, aws-ro
10. aws-ro, consul-template, nebula, slack-deployable, prometheus_targets, consul, slack-base-log, murron, aws-ro
11. aws-ro, consul-template, nebula, slack-deployable, prometheus_targets, consul, slack-base-log, slack-monitoring, slack-apache, php, aws-ro
12. certs, consul-template, nebula, slack-deployable, consul, certs
13. certs, consul-template, nebula, slack-deployable, consul, slack-base-log, slack-monitoring, certs
14. certs, consul-template, nebula, slack-deployable, consul, slack-base-log, slack-monitoring, slack-apache, certs
15. certs, consul-template, nebula, slack-deployable, prometheus_targets, consul, certs
16. certs, consul-template, nebula, slack-deployable, prometheus_targets, consul, slack-base-log, slack-monitoring, certs
17. certs, consul-template, nebula, slack-deployable, prometheus_targets, consul, slack-base-log, slack-monitoring, slack-apache, certs
18. consul, consul-template, nebula, slack-deployable, consul
19. consul, consul-template, nebula, slack-deployable, prometheus_targets, consul
20. consul, nebula, slack-deployable, consul
21. consul, nebula, slack-deployable, prometheus_targets, consul
22. consul, slack-base-log, murron, consul
23. consul, slack-base-log, murron, slack-deployable, consul
24. consul, slack-base-log, murron, slack-deployable, prometheus_targets, consul
25. consul, slack-base-log, slack-monitoring, consul
26. consul, slack-base-log, slack-monitoring, prometheus_targets, consul
27. consul, slack-base-log, slack-monitoring, prometheus_targets, nebula, slack-deployable, consul
28. consul, slack-base-log, slack-monitoring, prometheus_targets, slackops-tool, consul-template, nebula, slack-deployable, consul
29. consul, slack-base-log, slack-monitoring, prometheus_targets, slackops-tool, slack-deployable, consul
30. consul, slack-base-log, slack-monitoring, slack-apache, consul-template, nebula, slack-deployable, consul
31. consul, slack-base-log, slack-monitoring, slack-apache, consul-template, nebula, slack-deployable, prometheus_targets, consul
32. consul, slack-base-log, slack-monitoring, slackops-tool, consul-template, nebula, slack-deployable, consul
33. consul, slack-base-log, slack-monitoring, slackops-tool, consul-template, nebula, slack-deployable, prometheus_targets, consul
34. consul, slack-base-log, slack-monitoring, slackops-tool, slack-deployable, consul
35. consul, slack-base-log, slack-monitoring, slackops-tool, slack-deployable, prometheus_targets, consul
36. consul, slack-base-log, slack-monitoring-client, nebula, slack-deployable, consul
37. consul, slack-base-log, slack-monitoring-client, nebula, slack-deployable, prometheus_targets, consul
38. consul, slack-base-log, slack-monitoring-client, prometheus_targets, consul
39. consul, slack-base-log, slack-monitoring-client, prometheus_targets, nebula, slack-deployable, consul
40. consul, slack-base-log, slack-monitoring-client, prometheus_targets, slackops-tool, consul-template, nebula, slack-deployable, consul
41. consul, slack-base-log, slack-monitoring-client, prometheus_targets, slackops-tool, slack-deployable, consul
42. consul, slackops-tool, consul-template, nebula, slack-deployable, consul
43. consul, slackops-tool, consul-template, nebula, slack-deployable, prometheus_targets, consul
44. consul, slackops-tool, slack-deployable, consul
45. consul, slackops-tool, slack-deployable, prometheus_targets, consul
46. consul, what-happened, slack-deployable, consul
47. consul, what-happened, slack-deployable, prometheus_targets, consul
48. consul-template, nebula, slack-collectd, consul-template
49. consul-template, nebula, slack-deployable, consul-template
50. consul-template, nebula, slack-deployable, prometheus_targets, slackops-tool, consul-template
51. nebula, slack-deployable, prometheus_targets, nebula
52. prometheus_targets, slackops-tool, slack-deployable, prometheus_targets
53. slack-deployable, what-happened, slack-deployable
```

## Deployments

We are currently using `orchestrated-deploy` for Linux machines and `slack-cli-tools` for MacOS deployments.

### Linux

The pipeline to publish and promote new versions for Linux is: https://jenkins.tinyspeck.com/job/whisk

### MacOS

1. Run `./build.sh`
1. Copy the binaries generated in `./build/` to https://slack-github.com/slack/slack-cli-tools/tree/master/bin
1. Send a PR to `slack-cli-tools`

## Integrations
### Checkpoint
We currently have Whisk configured as a CI check in the following repos: 
* chef-repo: https://slack-github.com/slack/chef-repo/blob/f887f64cc4efebea7980848313ec5baaf3413bd2/slack.json#L115-L117

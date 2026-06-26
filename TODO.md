# Chronix TODO

_Working idea capture for future improvements and deferred enhancements._

## Product

- [ ] Build an epic testing platform for all of Chronix.
  This is the next major thing to tackle. Keep it at the top of the list so it can be the first item crossed off once work starts.

- [ ] Add a guided job wizard that can create a connection, action, and job in one flow.
  Let the user choose an existing connection or step through creating one if it does not exist yet. Then let them pick an existing action or create a new one before moving into job creation. Later, we can decide how deep the wizard should go, including whether each stage should use the existing full-screen forms or a more granular step-by-step guided flow.

- [ ] Design agent update notification levels and opt-in release lanes.
  Need a user-selectable model for which agent/server updates are surfaced. Current idea set includes stable maintenance releases, broader minor releases, and an opt-in prerelease lane between normal maintenance and general availability notifications.

- [ ] Explore support for multiple independent agent identities on one machine.
  Keep this as separate local agent identities or profiles rather than one agent process connected to multiple servers. Potential fit for a future advanced operations lane.

- [ ] Redesign activity logs, alerts, and reporting before the version 1 push.
  This needs a dedicated design pass. Alerts should represent important messages without becoming cluttered. Logging needs to be comprehensive enough to trust, but manageable enough to use without overwhelming the operator. Reporting has not even really been started yet, and it is a major product capability, so this entire area is especially critical.

  Open questions to resolve:
  How should high-frequency job activity be represented when jobs can run every minute?
  Do we log every run, summarize runs, or emphasize only failures and exceptions?
  Where should the cap, retention, and roll-up behavior live for noisy activity streams?
  How should run results, user activity, logins, configuration changes, and connection history be presented so they remain useful instead of becoming a wall of noise?

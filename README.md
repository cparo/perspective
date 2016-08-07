# Perspective

Graphing library for quality control in event-driven systems.

## MVP Implementation

https://www.youtube.com/watch?v=GqKOQ4X5KUM

This is obviously a work in progress - and subject to major changes in both
interface and implementation until this README is changed to say otherwise.

In its MVP form (as shown in the above talk), Perspective provides:
* A library of visualization-generation objects
* A set of command-line utilities for generating visualizations with those
  objects
* An HTTP service for generating visualizations with those objects for
  inclusion in web dashboards

It is not what I would call idiomatic Go (though I did learn a lot about the
preferred idioms for Go development through feedback on this and other projects
I did while learning the language). It is also not, in this MVP form, an ideal
design for the use cases it has in practice proven to be mosed used for...

## The Future

Recognizing that Perspective's actual usage has gravitated toward real-time
display of system behavior (as opposed to the ad-hoc display of behavior from
past time windows), the following changes are sensible:

* Make a server which can receive incremental updates (as could be obtained
  from the source of a recorded event, from a message-queue or distributed log
  system, or from an agent polling state as recorded in another system) - and
  append these to a persisted log and in-memory internal representation of
  these events and their state transitions.
* Move the actual graphical rendering process to the client. Where Go may have
  an advantage of JavaScript in raw computational efficiency, client-side JS
  code has the advantage of inherently scaling out linearly with the number of
  clients displaying Perspective's data. Also, this will allow us to stream
  incremental updates to these clients without a heavyweight polling-refresh
  cycle, giving clients the ability to map new data points into their
  visualizations as they are received.
* Graph labels can be rendered in the client in direct connection to the graphs
  themselves, removing a source of potential error or confusion which exists in
  the current system.
* A "snapshot" feature on the client would be useful for rendering a snapshot
  of current state of a visualization to a timestamp-named png file for sharing
  or copying into a document (as may be useful for post-mortem discussions).
* Configurable composited layers with configurable colors will be useful for a
  more flexible display of incoming data, making the tool more suitable for use
  in contexts outside of its original case as a complement for automated system
  monitoring in the management of DigitalOcean's cloud infrastructure.

For this change, an evaluation of using the D3 visualization framework vs. a
bespoke mechanism for rendering to a graph canvas would be appropriate before
commiting to one approach or the other. Generally speaking, D3 would be
favorable as a commonly-known visualization framework so long as performance
works out, so we will start there and then considering other options if
necessary for performance reasons.

Finally, while the current core scatter/heat-map visualization is appropriate
for the display of discrete recorded events, I have some thought on adapting
line-graph type displays to work better with the display of multiple continuous
or quasi-continuous trends (like CPU utilization or stock prices) in a compact
format which also gives information on the relative weight of these trends
overall (ex: the relative significance of price changes in members of a stock
portfilio based on their allocation weight, or the prevalence of a system
performance/utiization trend based on the number of workloads sampled as
members of that trend - as could be the case in tracking performance trends on
two versions of a platform during a migration from one to the other).

In a nutshell, this project should soon be revived for a substantial update
of its structure and functionality.

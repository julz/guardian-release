# Guardian - WORK IN PROGRESS!

Guardian is an extremely small and simple single-host manager for [Open Container Spec](http://github.com/opencontainers/specs) containers.
It contains the 'rundmc' runc wrapper (a lightweight wrapper around runC for managing multiple containers on a host) along with a simple volume/layered filesystem manager and a network component. 

## Why?

 - **for hackers**: Guardian is a super-small, super-simple non-opinionated wrapper around runC. That means it's extremely quick to get started, and easy to hack on.
 - **for container orchestrators**: Guardian is small and unopinionated, giving maximum flexibility to higher level orchestrators. Containers run independently and the Guardian daemon can restart and reattach to running containers without downtime.

## Components

The components are subdirectries of the root repository. These are:

 - **RunDMC**: A tiny wrapper for [runC](http://github.com/opencontainers/runc) to manage and reconnect to multiple containers.
 - **Garden Shed**: Downloads and manages rootfses and volumes.
 - **Kawasaki**: It's an amazing networker.
 - **Gardener**: Orchestrates the other components, implements the [Cloud Foundry Garden API](http://github.com/cloudfoundry-incubator/garden).
 - **Gardener's Question Time (GQT)**: A famous british radio show, and an integration test suite.

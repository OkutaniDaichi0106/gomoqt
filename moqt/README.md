# gomoqt

## Overview

This is an implementation of Media over QUIC Transfork.

## Specification

https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html

## Implementation Status
| Section                                      | Implemented        | Tested             |
| -------------------------------------------- | ------------------ | ------------------ |
| **2. Data Model**                            |                    |                    |
| 2.1. Frame                                   | :white_check_mark: | :white_check_mark: |
| 2.3. Group                                   | :white_check_mark: | :white_check_mark: |
| 2.4. Track                                   | :white_check_mark: | :white_check_mark: |
| 2.4.1. Track Naming and Scopes               | :construction:     | :x:                |
| 2.4.2. Scope                                 | :construction:     | :x:                |
| 2.4.3. Connection URL                        | :construction:     | :x:                |
| **3. Sessions**                              |                    |                    |
| 3.1. Session establishment                   | :white_check_mark: | :white_check_mark: |
| 3.1.1. WebTransport                          | :white_check_mark: | :white_check_mark: |
| 3.1.2. QUIC                                  | :white_check_mark: | :x:                |
| 3.2. Version and Extension Negotiation       | :white_check_mark: | :x:                |
| 3.3. Session initialization                  | :white_check_mark: | :white_check_mark: |
| 3.4. Stream Cancellation                     | :construction:     | :x:                |
| 3.5. Termination                             | :white_check_mark: | :white_check_mark: |
| 3.6. Migration                               | :construction:     | :x:                |
| **4. Data Transmittions**                    |                    |                    |
| 4.1 Track Priority Control                   | :construction:     | :x:                |
| 4.2 Group Order Control                      | :construction:     | :x:                |
| 4.3 Cache                                    | :construction:     | :x:                |
| **5. Relays**                                |                    |                    |
| 5.1. Subscriber Interactions                 | :white_check_mark: | :white_check_mark: |
| 5.1.1. Graceful Publisher Relay Switchover   | :x:                | :x:                |
| 5.2. Publisher Interactions                  | :white_check_mark: | :white_check_mark: |
| 5.2.1. Graceful Publisher Network Switchover | :x:                | :x:                |
| 5.2.2. Graceful Publisher Relay Switchover   | :x:                | :x:                |
| 5.3. Relay Object Handling                   | :construction:     | :x:                |
| **Control Streams**                          |                    |                    |
| 6.1. Session Stream                          | :white_check_mark: | :white_check_mark: |
| 6.2. Announce Stream                         | :white_check_mark: | :white_check_mark: |
| 6.3. Subscribe Stream                        | :white_check_mark: | :white_check_mark: |
| 6.4. Fetch Stream                            | :white_check_mark: | :white_check_mark: |
| 6.5. Info Stream                             | :white_check_mark: | :white_check_mark: |
| **Control Messages**                         |                    |                    |
| 6.1. Parameters                              | :construction:     | :white_check_mark: |
| 6.1.1. Version Specific Parameters           | :white_check_mark: | :white_check_mark: |
| 6.2. SESSION_CLIENT                          | :white_check_mark: | :white_check_mark: |
| 6.2. SESSION_SERVER                          | :white_check_mark: | :white_check_mark: |
| 6.2. SESSION_UPDATE                          | :white_check_mark: | :white_check_mark: |
| 6.2.1. Versions                              | :white_check_mark: | :white_check_mark: |
| 6.2.2. Setup Parameters                      | :white_check_mark: | :white_check_mark: |
| 6.3. ANNOUNCE_PLEASE                         | :white_check_mark: | :white_check_mark: |
| 6.4. ANNOUNCE                                | :white_check_mark: | :white_check_mark: |
| 6.5. SUBSCRIBE                               | :white_check_mark: | :white_check_mark: |
| 6.6. SUBSCRIBE_UPDATE                        | :white_check_mark: | :white_check_mark: |
| 6.7. SUBSCRIBE_GAP                           | :x:                | :x:                |
| 6.8. INFO                                    | :white_check_mark: | :white_check_mark: |
| 6.9. INFO_PLEASE                             | :white_check_mark: | :white_check_mark: |
| **Data Stream**                              |                    |                    |
| 7.3. Group Streams                           | :white_check_mark: | :white_check_mark: |
| **Data Message**                             |                    |                    |
| 7.2. GROUP                                   | :white_check_mark: | :white_check_mark: |
| 7.2. FRAME                                   | :white_check_mark: | :white_check_mark: |
| **Security Considerations**                  |                    |                    |
| 8.1. Resource Exhaustion                     | :x:                | :x:                |

## Interoperablity test
We haven't conducted interoperability testing with other implementations yet

## TODO
- Interoperability
- Scheduling
- LOC (Low Overhead Container)
- Common Catalog Format for moq
- sync.Pool

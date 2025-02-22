---
title: FAQ
layout: default
nav_order: 2
---

# Frequently Asked Questions

## Does **bitmagnet** download or distribute any illegal or copyright-infringing content?

No. **bitmagnet** does not download, store or distribute any content _at all_. It only downloads **metadata about** content. It may download **metadata about** illegal or copyright infringing content, and users should therefore exercise discretion in any magnet links they add to their BitTorrent client. **bitmagnet** attempts to detect and filter harmful content such as <abbr title="Child Sexual Abuse Material">CSAM</abbr> to avoid users having such undesirable metadata in their index.

## Should I use a VPN with **bitmagnet**?

It is recommended to use a VPN: **bitmagnet** may download **metadata about** illegal and copyrighted content. It is possible that rudimentary law enforcement and anti-piracy tracking tools would incorrectly flag this activity, although we've never heard about anyone getting into trouble for using this or similar metadata crawlers. Setting up a VPN is simple and cheap, and it's better to be safe than sorry. We are not affiliated with any VPN providers, but if you're unsure which provider to choose, we can recommend [Mullvad](https://mullvad.net/){:target="\_blank"}.

## What are the system requirements for **bitmagnet**?

As a rough guide, you should allow around 300MB RAM for BitMagnet, and at least 1GB RAM for the Postgres database. You should allow roughly 50GB of disk space per 10 million torrents, which should suffice for several months of crawling, however there is no upper limit to how many torrents might ultimately be crawled. The database will run fastest when it has plenty of RAM and a fast disk, preferably a SSD.

## I've started **bitmagnet** for the first time and am not seeing torrents right away, is something wrong?

If everything is working, **bitmagnet** should begin showing torrents in the web UI within a maximum of 10 minutes (which is its cache TTL). The round blue refresh button in the web UI is a cache buster - use it to see new torrent content in real time. Bear in mind that when a torrent is inserted into the database, a background queue job must run before it will become available in the UI. If you're importing thousands or millions of torrents, it might therefore take a while for everything to show. Here are some things to check if you're not seeing torrents:

- Press the round blue refresh button in the UI.
- Visit the metrics endpoint at `/metrics` and check the following metrics:
  - `bitmagnet_dht_crawler_persisted_total`: If you see a positive number for this, the DHT crawler is working and has found torrents.
  - If torrents are being persisted but you still don't see them in the UI, then check:`bitmagnet_queue_jobs_total{queue="process_torrent",status="processed"}`: If you see a positive number here, then the queue worker is running and processing jobs. If you see `status="failed"` or `status="retry"`, but no `status="processed"`, then something is wrong.
  - If no torrents are being persisted, check: `bitmagnet_dht_responder_query_success_total` and `bitmagnet_dht_responder_query_error_total`. Having some DHT query errors is completely normal, but if you see no successful queries then something is wrong.
- If the metrics confirm a problem, check the logs for errors.

## Why doesn't **bitmagnet** show me exactly how many torrents it has indexed?

Torrents are indexed to a Postgres database, and Postgres is notoriously slow in counting large numbers of rows. To provide acceptable performance, **bitmagnet** uses a strategy it calls a "budgeted count". This takes advantage of the fact that the Postgres query planner can provide an estimated count, along with the total cost of executing the count query. If the cost exceeds the budget, we return the estimate, and the UI will show an estimate symbol `~`. If the cost is within budget, we return the exact count. For large result sets, you will probably always be seeing an estimate.

## At what rate will **bitmagnet** crawl torrents from the DHT?

This will depend on a number of factors, including your hardware and network conditions, and your [`dht_crawler.scaling_factor` configuration](/setup/configuration.html). Typically it can be anything from 100 to 1,000 torrents per minute. Crawling is likely to slow down as your index grows larger, as it's more likely that any discovered torrent will already be in your index.

## How can I see exactly how many torrents **bitmagnet** has crawled in the current session?

Visit the metrics endpoint at `/metrics` and check the metric `bitmagnet_dht_crawler_persisted_total`. `{entity="Torrent"}` corresponds to newly crawled torrents, and `{entity="TorrentsTorrentSource"}` corresponds to torrents that were rediscovered and had their seeders/leechers count, and last-seen-on date updated.

## How are the seeders/leechers numbers determined for torrents crawled from the DHT?

The DHT crawler uses a [BEP33 scrape request](https://www.bittorrent.org/beps/bep_0033.html){:target="\_blank"} to provide a very rough estimate of the current seeders/leechers.

## How do I know if a torrent crawled by **bitmagnet** is being actively seeded, and that I'll be able to download it?

The short answer is you can't. The only way to know for sure is to add an info hash to your BitTorrent client. The seeders/leechers count provides an imperfect indicator of the torrent's health. In future **bitmagnet** may provide "decentralized tracker"-like features that would improve this.

## Can I ask **bitmagnet**'s DHT crawler to crawl specific hashes?

No. The DHT crawler works by sampling random info hashes from the network, and was not designed to locate specific hashes - it only crawls what it finds by chance. You can use the import [the `/import` endpoint](/tutorials/import.html) to import specific torrents, and additional methods (separate from the DHT crawler) may be added in future.

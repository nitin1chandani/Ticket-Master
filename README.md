# Ticket Master

## Why this exists

This project started from a simple pain point: ticket systems look easy until real users hit them at the same time.

If 500 people try to grab the same seats, naive code breaks. You get:
- double booking
- stale reservations
- weird booking states
- support nightmares

So this backend is built around that reality.

## What we are trying to solve

The focus here is not fancy architecture for show. The focus is correctness under pressure.

Main goals:
- no double-sell under concurrency
- temporary hold before payment
- clean expiry and release of abandoned reservations
- clear and predictable booking state transitions

## How booking works

We split booking into two clear steps:
- Reserve: hold tickets for a short window and create a pending booking
- Confirm: finalize reserved tickets and mark booking complete

If someone reserves and disappears, we do not keep inventory blocked forever.

A worker keeps checking expired reservations and does this:
- tickets back to available
- booking marked expired

## Tech choices

Under the hood:
- Postgres transactions for consistency
- Redis for short-lived reservation behavior
- background worker for expiry cleanup

Code layout is simple on purpose:
- handlers are thin
- services carry the business flow
- repositories handle DB operations



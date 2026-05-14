#!/bin/bash
psql $DATABASE_URL -c "\d invoices"

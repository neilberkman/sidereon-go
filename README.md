# sidereon-go

Construction stopped at the mandatory C ABI coverage gate. The public
`sidereon-c` v1.1.1 ABI cannot express every documented `sidereon-python`
v1.1.1 capability, and the authoritative interface brief prohibits an ABI
fork or a pure-Go numeric/parser workaround.

No Go module, implementation, bundled archive, test suite, tag, or release
artifact has been created. See the audited [coverage map](audit/COVERAGE_MAP.md),
[evidence](audit/AUDIT_EVIDENCE.md), and independent
[parity review](audit/PARITY_REVIEW.md).

Work may resume only after a same-version public `sidereon-c` release exposes
equivalent ABI routes for all required documented sibling capabilities, or the
authoritative scope explicitly changes and the coverage gate is rerun.

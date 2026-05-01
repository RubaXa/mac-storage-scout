# Contract Lint Report

- scanned_entities: 74
- violations: 0
- errors: 0
- warnings: 0

| Kind | Entity | File | Line | Status | Required Tags | Present Tags | Missing Required | Findings |
|---|---|---|---:|---|---|---|---|---|
| func | main | cmd/mss-contract-lint/main.go | 15 | ok | purpose, consumer | consumer, purpose | - | - |
| func | printText | cmd/mss-contract-lint/main.go | 51 | ok | purpose, consumer, param | consumer, param, purpose, returns | - | - |
| func | printMarkdown | cmd/mss-contract-lint/main.go | 93 | ok | purpose, consumer, param | consumer, param, purpose, returns | - | - |
| func | main | cmd/mss/main.go | 28 | ok | purpose, consumer | consumer, purpose | - | - |
| func | printUsage | cmd/mss/main.go | 49 | ok | purpose, consumer | consumer, purpose | - | - |
| func | runScan | cmd/mss/main.go | 60 | ok | purpose, consumer, param | consumer, param, purpose | - | - |
| func | runDelete | cmd/mss/main.go | 148 | ok | purpose, consumer, param | consumer, param, purpose | - | - |
| func | guardDeletePath | cmd/mss/main.go | 215 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | isProtectedDeletePath | cmd/mss/main.go | 234 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | validateTopN | cmd/mss/main.go | 254 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | measurePath | cmd/mss/main.go | 267 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | expandPath | cmd/mss/main.go | 304 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| type | Config | internal/contractlint/linter.go | 19 | ok | purpose, consumer | consumer, purpose | - | - |
| type | EntityKind | internal/contractlint/linter.go | 29 | ok | purpose, consumer | consumer, purpose | - | - |
| type | Entity | internal/contractlint/linter.go | 42 | ok | purpose, consumer | consumer, purpose | - | - |
| type | Severity | internal/contractlint/linter.go | 56 | ok | purpose, consumer | consumer, purpose | - | - |
| type | Finding | internal/contractlint/linter.go | 67 | ok | purpose, consumer | consumer, purpose | - | - |
| type | EntityReport | internal/contractlint/linter.go | 77 | ok | purpose, consumer | consumer, purpose | - | - |
| type | Report | internal/contractlint/linter.go | 90 | ok | purpose, consumer | consumer, purpose | - | - |
| func | Run | internal/contractlint/linter.go | 106 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | collectEntities | internal/contractlint/linter.go | 169 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | checkEntity | internal/contractlint/linter.go | 291 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | fieldsCount | internal/contractlint/linter.go | 348 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | receiverName | internal/contractlint/linter.go | 369 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | commentText | internal/contractlint/linter.go | 391 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | extractTags | internal/contractlint/linter.go | 404 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns, tag | - | - |
| func | sortedTagNames | internal/contractlint/linter.go | 425 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | isExportedEntity | internal/contractlint/linter.go | 443 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| type | MssTreeAggregatorAdapter | internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go | 17 | ok | purpose, consumer | consumer, implements, purpose | - | - |
| method | MssTreeAggregatorAdapter.BuildTree | internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go | 27 | ok | purpose, consumer, param, returns | consumer, param, post, pre, purpose, returns, see | - | - |
| func | dedupChildren | internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go | 104 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | recomputeSizes | internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go | 123 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | sortNodes | internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go | 143 | ok | purpose, consumer, param | consumer, param, purpose | - | - |
| func | applyThreshold | internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go | 159 | ok | purpose, consumer, param | consumer, param, purpose | - | - |
| func | collectTypesFromNodes | internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go | 234 | ok | purpose, consumer, param | consumer, param, purpose | - | - |
| type | MssGoFsWalkerAdapter | internal/mss/adapters/fs/mss_go_fs_walker_adapter.go | 23 | ok | purpose, consumer | consumer, implements, invariant, purpose | - | - |
| type | mssQueueItem | internal/mss/adapters/fs/mss_go_fs_walker_adapter.go | 29 | ok | purpose, consumer | consumer, purpose | - | - |
| method | MssGoFsWalkerAdapter.Walk | internal/mss/adapters/fs/mss_go_fs_walker_adapter.go | 43 | ok | purpose, consumer, param, returns | consumer, param, post, pre, purpose, returns, see | - | - |
| method | MssGoFsWalkerAdapter.walkPath | internal/mss/adapters/fs/mss_go_fs_walker_adapter.go | 124 | ok | purpose, consumer, param | consumer, param, purpose | - | - |
| func | mssExtOf | internal/mss/adapters/fs/mss_go_fs_walker_adapter.go | 230 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | mssExpandPath | internal/mss/adapters/fs/mss_go_fs_walker_adapter.go | 244 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | mssSizeFromStat | internal/mss/adapters/fs/mss_sys_stat_adapter.go | 17 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| type | MssAnsiProgressAdapter | internal/mss/adapters/progress/mss_ansi_progress_adapter.go | 19 | ok | purpose, consumer | consumer, implements, purpose | - | - |
| method | MssAnsiProgressAdapter.Start | internal/mss/adapters/progress/mss_ansi_progress_adapter.go | 27 | ok | purpose, consumer, param, returns | consumer, param, post, purpose, returns, see | - | - |
| func | mssIsTTY | internal/mss/adapters/progress/mss_ansi_progress_adapter.go | 76 | ok | purpose, consumer, returns | consumer, purpose, returns | - | - |
| type | MssTreeTextReportAdapter | internal/mss/adapters/report/mss_tree_text_report_adapter.go | 19 | ok | purpose, consumer | consumer, implements, invariant, purpose | - | - |
| method | MssTreeTextReportAdapter.Render | internal/mss/adapters/report/mss_tree_text_report_adapter.go | 30 | ok | purpose, consumer, param, returns | consumer, param, post, pre, purpose, returns, see | - | - |
| type | renderStyle | internal/mss/adapters/report/mss_tree_text_report_adapter.go | 69 | ok | purpose, consumer | consumer, purpose | - | - |
| method | MssTreeTextReportAdapter.printNodeChildren | internal/mss/adapters/report/mss_tree_text_report_adapter.go | 79 | ok | purpose, consumer, param | consumer, param, purpose | - | - |
| func | padRight | internal/mss/adapters/report/mss_tree_text_report_adapter.go | 178 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| type | MssScanOrchestrator | internal/mss/app/mss_scan_orchestrator.go | 21 | ok | purpose, consumer | consumer, implements, invariant, purpose | - | - |
| method | MssScanOrchestrator.Run | internal/mss/app/mss_scan_orchestrator.go | 37 | ok | purpose, consumer, param, returns | consumer, param, post, pre, purpose, returns, see | - | - |
| func | mssValidateOrchestratorConfig | internal/mss/app/mss_scan_orchestrator.go | 99 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | MssHumanBytes | internal/mss/domain/mss_format.go | 13 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| func | MssParseBytes | internal/mss/domain/mss_threshold.go | 19 | ok | purpose, consumer, param, returns | consumer, param, post, pre, purpose, returns | - | - |
| type | MssSizeMode | internal/mss/domain/mss_types.go | 11 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssScanConfig | internal/mss/domain/mss_types.go | 24 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssCounters | internal/mss/domain/mss_types.go | 39 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssEntryKind | internal/mss/domain/mss_types.go | 52 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssWalkEntry | internal/mss/domain/mss_types.go | 63 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssWalkEvent | internal/mss/domain/mss_types.go | 76 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssExtStat | internal/mss/domain/mss_types.go | 85 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssOtherBucket | internal/mss/domain/mss_types.go | 95 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssNode | internal/mss/domain/mss_types.go | 110 | ok | purpose, consumer | consumer, purpose | - | - |
| type | MssFilesystemWalkerPort | internal/mss/ports/mss_filesystem_walker_port.go | 14 | ok | purpose, consumer | consumer, purpose | - | - |
| interface-method | MssFilesystemWalkerPort.Walk | internal/mss/ports/mss_filesystem_walker_port.go | 23 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| type | MssNodeAggregatorPort | internal/mss/ports/mss_node_aggregator_port.go | 11 | ok | purpose, consumer | consumer, purpose | - | - |
| interface-method | MssNodeAggregatorPort.BuildTree | internal/mss/ports/mss_node_aggregator_port.go | 21 | ok | purpose, consumer, param, returns | consumer, param, post, pre, purpose, returns | - | - |
| type | MssProgressEmitterPort | internal/mss/ports/mss_progress_emitter_port.go | 14 | ok | purpose, consumer | consumer, purpose | - | - |
| interface-method | MssProgressEmitterPort.Start | internal/mss/ports/mss_progress_emitter_port.go | 21 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| type | MssReportComposerPort | internal/mss/ports/mss_report_composer_port.go | 14 | ok | purpose, consumer | consumer, purpose | - | - |
| interface-method | MssReportComposerPort.Render | internal/mss/ports/mss_report_composer_port.go | 23 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |
| type | MssScanOrchestratorPort | internal/mss/ports/mss_scan_orchestrator_port.go | 14 | ok | purpose, consumer | consumer, purpose | - | - |
| interface-method | MssScanOrchestratorPort.Run | internal/mss/ports/mss_scan_orchestrator_port.go | 22 | ok | purpose, consumer, param, returns | consumer, param, purpose, returns | - | - |

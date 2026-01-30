# Implementation Plan: Syncp Documentation Support

## Overview

This implementation plan extends the NATS Documentation MCP Server to support dual documentation sources (NATS and Synadia Control Plane). The approach follows an incremental strategy: first implement the query classifier, then extend the index manager for dual indices, add the search orchestrator, extend the fetcher, update configuration, and finally integrate everything with the MCP server tools. Each major component includes property-based tests to validate correctness properties from the design.

## Tasks

- [ ] 1. Implement Query Classifier with keyword-based classification
  - [ ] 1.1 Create classifier package with DocumentationSource enum and Classifier interface
    - Define DocumentationSource constants (SourceNATS, SourceSyncp, SourceBoth)
    - Define Classifier interface with Classify(query string) method
    - Create KeywordClassifier struct with keyword lists
    - _Requirements: 3.1, 3.5_
  
  - [ ] 1.2 Implement keyword-based classification algorithm
    - Normalize query to lowercase
    - Tokenize query into words
    - Count matches against syncp and NATS keyword lists
    - Apply classification rules (syncp-only → SourceSyncp, nats-only → SourceNATS, both/neither → SourceBoth)
    - _Requirements: 3.2, 3.3, 3.4_
  
  - [ ] 1.3 Write property test for classification determinism
    - **Property 3: Classification Determinism**
    - **Validates: Requirements 3.1**
  
  - [ ] 1.4 Write property test for keyword-based classification correctness
    - **Property 4: Keyword-Based Classification Correctness**
    - **Validates: Requirements 3.2, 3.3, 3.4**
  
  - [ ] 1.5 Write property test for classification configuration sensitivity
    - **Property 5: Classification Configuration Sensitivity**
    - **Validates: Requirements 3.5**
  
  - [ ] 1.6 Write unit tests for classification edge cases
    - Test empty query string
    - Test query with no keyword matches
    - Test query with special characters
    - _Requirements: 3.4_

- [ ] 2. Extend Index Manager for dual documentation sources
  - [ ] 2.1 Create Index Manager with separate NATS and syncp indices
    - Create Manager struct with natsIndex and syncpIndex fields
    - Implement IndexNATS and IndexSyncp methods
    - Implement GetNATSIndex and GetSyncpIndex accessors
    - Implement Stats method for memory reporting
    - _Requirements: 2.1, 2.3, 9.3_
  
  - [ ] 2.2 Write property test for index independence
    - **Property 1: Index Independence**
    - **Validates: Requirements 2.1**
  
  - [ ] 2.3 Write unit tests for index manager operations
    - Test indexing NATS docs only
    - Test indexing syncp docs only
    - Test indexing both sources
    - Test stats reporting
    - _Requirements: 2.1, 2.3, 9.3_

- [ ] 3. Implement Search Orchestrator for multi-source search
  - [ ] 3.1 Create Orchestrator struct with dual indices and classifier
    - Define SearchResult struct with Source field
    - Create Orchestrator with natsIndex, syncpIndex, and classifier
    - Implement NewOrchestrator constructor
    - _Requirements: 4.1, 4.2, 4.3, 7.1_
  
  - [ ] 3.2 Implement Search method with classification-based routing
    - Call classifier.Classify(query) to determine source
    - Route to appropriate index based on classification
    - Annotate results with source metadata
    - Merge and sort results when searching both indices
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  
  - [ ] 3.3 Write property test for source-specific routing
    - **Property 6: Source-Specific Routing**
    - **Validates: Requirements 4.1, 4.2**
  
  - [ ] 3.4 Write property test for merged results completeness
    - **Property 7: Merged Results Completeness**
    - **Validates: Requirements 4.3**
  
  - [ ] 3.5 Write property test for result source annotation
    - **Property 8: Result Source Annotation**
    - **Validates: Requirements 4.4, 7.1, 7.2**
  
  - [ ] 3.6 Write property test for score-based sorting invariant
    - **Property 9: Score-Based Sorting Invariant**
    - **Validates: Requirements 4.5**
  
  - [ ] 3.7 Write unit tests for search orchestrator edge cases
    - Test search with empty NATS index
    - Test search with empty syncp index
    - Test search with both indices empty
    - _Requirements: 4.3_

- [ ] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 5. Extend Documentation Fetcher for syncp source
  - [ ] 5.1 Create MultiSourceFetcher with NATS and syncp configurations
    - Define Config struct for fetcher configuration
    - Create MultiSourceFetcher with natsConfig and syncpConfig
    - Implement NewMultiSourceFetcher constructor
    - _Requirements: 1.1, 1.2_
  
  - [ ] 5.2 Implement FetchSyncp method with retry logic
    - Reuse existing HTTP client and retry logic
    - Fetch syncp documentation from docs.synadia.com/control-plane
    - Apply same rate limiting as NATS fetcher
    - Handle errors gracefully (log and continue)
    - _Requirements: 1.1, 1.2, 1.3_
  
  - [ ] 5.3 Write property test for retry logic consistency
    - **Property 13: Retry Logic Consistency**
    - **Validates: Requirements 1.2**
  
  - [ ] 5.4 Write unit tests for syncp fetch error handling
    - Test network errors (connection refused, timeout)
    - Test HTTP errors (404, 500)
    - Test graceful degradation to NATS-only mode
    - _Requirements: 1.3_

- [ ] 6. Extend Parser for consistent multi-source parsing
  - [ ] 6.1 Verify parser works with syncp documentation structure
    - Test parser with sample syncp HTML pages
    - Ensure ParsedDocument structure is source-agnostic
    - Add source field to Document struct if needed
    - _Requirements: 2.2_
  
  - [ ] 6.2 Write property test for parser consistency
    - **Property 2: Parser Consistency**
    - **Validates: Requirements 2.2**
  
  - [ ] 6.3 Write unit tests for syncp-specific parsing scenarios
    - Test syncp page structure
    - Test syncp navigation links
    - Test syncp content extraction
    - _Requirements: 2.2_

- [ ] 7. Extend Configuration Manager for syncp settings
  - [ ] 7.1 Add syncp configuration fields to config schema
    - Add syncp.enabled boolean field
    - Add syncp.base_url string field
    - Add syncp.fetch_timeout duration field
    - Add classification.syncp_keywords string array
    - Add classification.nats_keywords string array
    - _Requirements: 6.1, 6.2, 6.5_
  
  - [ ] 7.2 Implement configuration validation for syncp URLs
    - Validate URL format (must be valid HTTP/HTTPS URL)
    - Validate URL is not empty when syncp is enabled
    - Return validation errors for invalid configurations
    - _Requirements: 6.4_
  
  - [ ] 7.3 Implement default values for syncp configuration
    - Default syncp.enabled = false
    - Default syncp.base_url = "https://docs.synadia.com/control-plane"
    - Default keyword lists (syncp and NATS keywords from design)
    - _Requirements: 6.5_
  
  - [ ] 7.4 Write property test for URL validation
    - **Property 11: URL Validation**
    - **Validates: Requirements 6.4**
  
  - [ ] 7.5 Write unit tests for configuration backward compatibility
    - Test existing config files parse correctly
    - Test missing syncp config defaults to NATS-only mode
    - Test syncp disabled mode
    - _Requirements: 5.3, 5.4, 6.3_

- [ ] 8. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 9. Update MCP Server tool handlers for dual-source support
  - [ ] 9.1 Integrate Search Orchestrator into search_nats_docs tool handler
    - Replace direct index queries with orchestrator.Search calls
    - Preserve existing tool interface (no breaking changes)
    - Include source metadata in tool responses
    - _Requirements: 5.1, 7.1, 7.3_
  
  - [ ] 9.2 Update retrieve_nats_doc tool handler for source metadata
    - Add source field to document retrieval responses
    - Preserve all existing metadata fields
    - _Requirements: 7.2, 7.3_
  
  - [ ] 9.3 Write property test for result metadata preservation
    - **Property 12: Result Metadata Preservation**
    - **Validates: Requirements 7.3**
  
  - [ ] 9.4 Write unit tests for tool handler integration
    - Test search_nats_docs with NATS-only queries
    - Test search_nats_docs with syncp-only queries
    - Test search_nats_docs with mixed queries
    - Test retrieve_nats_doc with source metadata
    - _Requirements: 5.1, 7.1, 7.2_

- [ ] 10. Implement logging for query classification
  - [ ] 10.1 Add debug logging to Query Classifier
    - Log classification decision (NATS, Syncp, or Both)
    - Log matched keywords for each source
    - Log ambiguous classifications
    - _Requirements: 8.1, 8.2, 8.3_
  
  - [ ] 10.2 Write unit tests for classification logging
    - Test log entries are created for classifications
    - Test log entries contain matched keywords
    - Test ambiguous query logging
    - _Requirements: 8.1, 8.2, 8.3_

- [ ] 11. Update server startup to initialize dual-source system
  - [ ] 11.1 Modify main.go to create MultiSourceFetcher
    - Load syncp configuration from config file
    - Create MultiSourceFetcher with NATS and syncp configs
    - Fetch both documentation sources at startup
    - _Requirements: 1.1, 6.1, 6.2_
  
  - [ ] 11.2 Initialize Index Manager with dual indices
    - Create Index Manager
    - Index NATS documentation
    - Index syncp documentation (if enabled)
    - Report memory statistics at startup
    - _Requirements: 2.1, 9.3_
  
  - [ ] 11.3 Initialize Query Classifier with keyword configuration
    - Load keyword lists from configuration
    - Create KeywordClassifier with loaded keywords
    - _Requirements: 3.5, 6.1_
  
  - [ ] 11.4 Initialize Search Orchestrator with all components
    - Create Orchestrator with dual indices and classifier
    - Pass orchestrator to MCP tool handlers
    - _Requirements: 4.1, 4.2, 4.3_
  
  - [ ] 11.5 Write integration test for full startup flow
    - Test startup with syncp enabled
    - Test startup with syncp disabled
    - Test startup with syncp fetch failure
    - _Requirements: 1.1, 1.3, 5.2, 6.3_

- [ ] 12. Implement backward compatibility validation
  - [ ] 12.1 Write property test for backward compatibility with syncp disabled
    - **Property 10: Backward Compatibility with Syncp Disabled**
    - **Validates: Requirements 5.2, 6.3**
  
  - [ ] 12.2 Write integration tests for backward compatibility
    - Test existing configuration files work unchanged
    - Test NATS-only mode produces same results as original
    - Test tool interfaces unchanged
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 13. Checkpoint - Run full integration test suite
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 14. Update configuration example file
  - [ ] 14.1 Add syncp configuration section to config.example.yaml
    - Add syncp.enabled, syncp.base_url, syncp.fetch_timeout
    - Add classification.syncp_keywords and classification.nats_keywords
    - Include comments explaining each field
    - _Requirements: 6.1, 6.2, 6.5_

- [ ] 15. Update documentation
  - [ ] 15.1 Update README.md with syncp support information
    - Document syncp configuration options
    - Explain query classification behavior
    - Provide examples of NATS vs syncp queries
    - _Requirements: 3.1, 6.1, 6.2_
  
  - [ ] 15.2 Update CONTRIBUTING.md with dual-source testing guidance
    - Document property test requirements
    - Explain classification testing approach
    - Provide examples of testing dual-source features
    - _Requirements: 2.1, 3.1_

- [ ] 16. Final validation and testing
  - [ ] 16.1 Run all property tests with 100+ iterations
    - Verify all 13 properties pass
    - Check for edge cases discovered by property tests
    - _Requirements: All_
  
  - [ ] 16.2 Run full integration test suite
    - Test end-to-end flows for all classification scenarios
    - Test error handling and graceful degradation
    - Test backward compatibility
    - _Requirements: All_
  
  - [ ] 16.3 Perform manual testing with real documentation sources
    - Test with live docs.nats.io
    - Test with live docs.synadia.com/control-plane
    - Verify classification accuracy with real queries
    - _Requirements: 1.1, 3.2, 3.3, 3.4_

## Notes

- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at key milestones
- Property tests validate universal correctness properties (13 total)
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end flows and component interactions
- Follow TDD principles: write tests first, then implement to make them pass

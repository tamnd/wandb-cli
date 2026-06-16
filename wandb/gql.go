package wandb

// GraphQL query strings for the W&B public API.
// These target only fields available without authentication.

const queryPublicViews = `
query PublicViews($type: String, $first: Int, $after: String) {
  publicViews(type: $type, first: $first, after: $after) {
    edges {
      node {
        id
        displayName
        description
        createdAt
        updatedAt
        entityName
        projectName
        coverUrl
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`

const queryFeaturedReports = `{
  featuredReports {
    id
    displayName
    description
    createdAt
    updatedAt
    entityName
    projectName
    coverUrl
  }
}`

const queryReportSearch = `
query ReportSearch($query: String!) {
  reportSearch(query: $query) {
    edges {
      node {
        id
        displayName
        description
        createdAt
        updatedAt
        entityName
        projectName
      }
    }
  }
}`

const queryEntity = `
query Entity($name: String!) {
  entity(name: $name) {
    id
    name
    isTeam
    memberCount
    photoUrl
    projectCount
    createdAt
  }
}`

const queryProject = `
query Project($entityName: String!, $name: String!) {
  project(entityName: $entityName, name: $name) {
    id
    name
    entityName
    description
    isPublic
    runCount
    createdAt
  }
}`

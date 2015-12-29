Feature:
  In order to have an up-to-date HTTP front-end for my k8s services
  As a developer
  I want this application to updates the Vamp routes

  Scenario: Route a created service
    When a k8s service named "app" is created in the namespace "qwerty"
    Then the vamp service "app.qwerty" should be created

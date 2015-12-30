Feature:
  In order to have an up-to-date HTTP front-end for my k8s services
  As a developer
  I want this application to updates the Vamp routes

  Scenario: Route a created service
    When a k8s service named "app" is created in the namespace "qwerty"
    Then the vamp route "http" should be created

  Scenario: VR Filter & Service created
    Given a vamp route named "http" already exists
    When a k8s service named "app" is created in the namespace "qwerty" with the IP "1.2.3.4"
    Then the vamp filter named "app.qwerty" should be created
    And the vamp service "app.qwerty" should be created
    And the vamp service "app.qwerty" should only contain the backend "1.2.3.4"

  Scenario: Updates the backend address if the service changes
    Given a vamp route named "http" already exists
    And a k8s service named "app" is created in the namespace "qwerty" with the IP "1.2.3.4"
    When a k8s service named "app" is updated in the namespace "qwerty" with the IP "2.3.4.5"
    Then the vamp service "app.qwerty" should only contain the backend "2.3.4.5"

  Scenario: Should not update the service if nothing changed
    Given a vamp route named "http" already exists
    When a k8s service named "app" is created in the namespace "qwerty" with the IP "1.2.3.4"
    Then the vamp route should be updated
    When a k8s service named "app" is updated in the namespace "qwerty" with the IP "1.2.3.4"
    Then the vamp route should not be updated

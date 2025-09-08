# Practical: Automated End-to-End Testing with Playwright

This practical exercise demonstrates how automated end-to-end testing can be integrated into a modern development workflow using Playwright, GitHub Actions, and Slack for real-time notifications.

## Objectives

- Implement robust E2E test scenarios with Playwright  
- Configure CI/CD pipelines using GitHub Actions  
- Deliver instant build/test updates through Slack integration  

## Key Highlights

### Playwright Testing
End-to-end test scripts were created to validate critical user journeys. All test cases executed successfully, confirming the correctness of UI flows and backend integration.  

![alt text](<assets/Screenshot 2025-09-07 214221.png>)

### GitHub Actions CI Pipeline
A GitHub Actions workflow was configured to automatically run the Playwright test suite on every code push. The pipeline executed flawlessly, ensuring consistency and early detection of errors.  

![alt text](<Screenshot 2025-09-07 232013.png>)

### Slack Notifications
Integration with Slack provided instant feedback by delivering success/failure messages to a dedicated channel whenever the CI pipeline was triggered. This keeps the team updated in real time without checking logs manually.  

![alt text](<Screenshot 2025-09-07 231946.png>)

## Outcome

This practical illustrates the power of combining Playwright testing with continuous integration and team collaboration tools. The automated pipeline reduces manual effort, increases reliability, and improves communication across the team.


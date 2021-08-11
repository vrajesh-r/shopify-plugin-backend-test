# Merchant Onboarding for Bread - Shopify Payments Solution

### App Installation Url
`https://shopify.getbread.com/install?shop=<shopify_shop_name>`

### Cart Button
Add this button to your cart page markup, and ensure that class & id stay the same.

`<div class="bread-integration-btn" id="bread-checkout-btn"></div>`

### Product Detail Button
Add this button to your product detail page markup, and ensure that class & id stay the same.

` <div class="bread-integration-btn-product" id="bread-checkout-btn-product></div>`

### Settings

Settings available to customize Bread - Shopify integration.

| Name | Purpose | Notes |
|:---------|----------:|:----------|
| API Key | Bread Identifier | Value for this varies between production & sandbox environments |
| API Secret | Bread secret credential | Value for field varies between production & sandbox environments |
| Custom CSS | Used to style Bread buttons on product detail & cart pages | Do not wrap in quotes |
| Auto-Settle Payments | When activated, payments will auto settle on checkout | Only activate when you can ensure ability to fulfill order |
| Save New Customers | When activated, Bread will save customer to your Shopify store after a successful checkout | - |
| Product DIV Acts as Label | When activated, product detail button will act as a label for customers logged in with Bread | - |
| Environment | Check to go live, uncheck to operate in sandbox | Use sandbox for development and testing, production when you publish feature |


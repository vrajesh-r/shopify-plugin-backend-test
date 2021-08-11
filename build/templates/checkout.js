(function(window, document) {
  // Go templated variables
  var GATEWAY_HOST = '{{ .MiltonHost }}';
  var BREAD_JS = '{{ .BreadJS }}';
  var BREAD_API_KEY = '{{ .ApiKey }}';
  var ENABLE_EMBEDDED_CHECKOUT = {{ .EnableCheckout }};
  var HEALTH_CARE = {{ .HealthcareMode }};
  var HEALTH_CARE_DISCLOSURE = 'This application is powered by Bread. Bread is not affiliated with this merchant. Any information you provide here will be received by Bread only, and Bread’s privacy policy will govern the use of this information.';
  var TX_RECORD_URL = GATEWAY_HOST + '/gateway/checkout/plus/record';
  var TARGETED_FINANCING = {{ .TargetedFinancing }};
  var TARGETED_FINANCING_THRESHOLD = {{ .TargetedFinancingThreshold }};
  var TARGETED_FINANCING_PROGRAM_ID = '{{ .TargetedFinancingProgramID }}';
  var BREAD_FORM_ID = 'bread-checkout-form';

  // Bread Gateways
  var BREAD_GATEWAY_KEY = {
    'Bread': true,
    'Bread (Staging)': true,
    'Bread (Sandbox)': true,
    'Bread (Development)': true
  };

  // Document state variables
  var BILLING_INPUT_CHANGE = 0;

  var ERROR_LABEL = '[Bread-Shopify] ';

  var selectors = window.BreadShopifyPlus.selectors || {};

  // DOM - gateway, billing input, shipping input selectors
  var ALL_GATEWAY_SELECTOR = selectors.allGateways || '[data-gateway-group]';
  var PLACE_ORDER_BTN_SELECTOR = selectors.placeOrderBtn || '.shown-if-js button[data-trekkie-id="complete_order_button"]';
  var BILLING_ADDRESS_FORM_SELECTOR = selectors.billingAddressForm || '[data-billing-address]';
  var BILLING_INPUT_SELECTOR = selectors.billingInput || '[data-backup="different_billing_address_true"]';
  var SHIPPING_INPUT_SELECTOR = selectors.shippingInput || '[data-backup="different_billing_address_false"]';
  var DATA_POLL_SELECTOR = selectors.dataPollRefresh || '[data-poll-refresh]';
  var ORDER_TOTAL_DATA_ATTRIBUTE = selectors.orderTotalDataAttribute || 'data-checkout-payment-due-target';
  var ORDER_TOTAL_SELECTOR = selectors.orderTotal || '[' + ORDER_TOTAL_DATA_ATTRIBUTE + ']';
  var TAX_TOTAL_DATA_ATTRIBUTE = selectors.taxTotalDataAttribute || 'data-checkout-total-taxes-target';
  var TAX_TOTAL_SELECTOR = selectors.taxTotal || '[' + TAX_TOTAL_DATA_ATTRIBUTE + ']';

  // DOM - checkout selector, id helper functions
  var CHECKOUT_NODE_SELECTOR = function ($gateway) {
    return '[data-subfields-for-gateway="' + $gateway.getAttribute('data-select-gateway') + '"]';
  };
  var CHECKOUT_NODE_ID_CREATOR = function ($gateway) {
    return 'payment-gateway-subfields-' + $gateway.getAttribute('data-select-gateway');
  };

  // Return query selector string for Shopify "Place order" button
  var GET_PLACE_ORDER_SELECTOR = function () {
    if (document.querySelector('.shown-if-js button[data-trekkie-id="complete_order_button"]')) {
      return '.shown-if-js button[data-trekkie-id="complete_order_button"]';
    } else if (document.querySelector('.shown-if-js button[type="submit"]')) {
      return '.shown-if-js button[type="submit"]';
    } else {
      return 'button[type="submit"]';
    }
  }

  // Adding this function to handle Shopify's changes to checkout
  var BILLING_STATE_OPTION_VALUE = function($state, billingContactState) {
    var $stateOption = $state.querySelector('option[value="' + billingContactState + '"]');
    return $stateOption ? $stateOption.value : $state.querySelector('[data-code="' + billingContactState + '"]').getAttribute('value');
  };

  // DOM - billing address selectors
  var BILLING_ADDRESS_FIRST_NAME = selectors.billingFirstName || '[data-address-field="first_name"]';
  var BILLING_ADDRESS_LAST_NAME = selectors.billingLastName || '[data-address-field="last_name"]';
  var BILLING_ADDRESS_ADDRESS_1 = selectors.billingAddress1 || '[data-address-field="address1"]';
  var BILLING_ADDRESS_ADDRESS_2 = selectors.billingAddress2 || '[data-address-field="address2"]';
  var BILLING_ADDRESS_CITY = selectors.billingCity || '[data-address-field="city"]';
  var BILLING_ADDRESS_COUNTRY = selectors.billingCountry || '[data-address-field="country"]';
  var BILLING_ADDRESS_PROVINCE = selectors.billingProvince || '[data-address-field="province"]';
  var BILLING_ADDRESS_ZIP = selectors.billingZip || '[data-address-field="zip"]';
  var BILLING_ADDRESS_PHONE = selectors.billingPhone || '[data-address-field="phone"]';

  var shouldExit = function() {
    if (window.Shopify.Checkout.step !== 'payment_method') {
      return true;
    }
    if (!ENABLE_EMBEDDED_CHECKOUT) {
      console.warn('[Bread-Shopify] Embedded checkout disabled in gateway settings');
      return true;
    }
    if (BREAD_API_KEY === '') {
      console.warn('[Bread-Shopify] No API key provided, exiting');
      return true;
    }
    return false;
  };

  var validateOptsItems = function() {
    window.BreadShopifyPlus.opts.items.forEach(function(item) {
      item.name = item.name !== '' ? item.name : 'missing';
      item.sku = item.sku !== '' ? item.sku : 'missing';
      item.imageUrl = item.imageUrl !== '' ? item.imageUrl : 'missing';
      item.detailUrl = formatDetailUrl(item.detailUrl);
    });
  };

  // Abort checkout.js if settings are invalid
  if (shouldExit()) {
    return;
  }
  loadCheckoutJS(BREAD_JS, BREAD_API_KEY);

  function formatDetailUrl(detailUrl) {
    if (detailUrl === '') {
      return 'missing';
    }

    if (detailUrl.toLowerCase().indexOf("/products") === 0) {
      return window.location.protocol + "//" + window.location.hostname + detailUrl;
    }

    return detailUrl;
  }

  function pollCheckoutElement() {
    // Check DOM for data-poll-refresh element which indicates that Shopify checkout form is not ready
    if (document.querySelectorAll(DATA_POLL_SELECTOR).length === 0) {
      initCheckout(); // Skip to checkout if not found
      return;
    }
    // Continue polling for polling elements if found
    var clear = setInterval(function() {
      // Interval should continue until polling elements are not found
      if (document.querySelectorAll(DATA_POLL_SELECTOR).length > 0) {
        return;
      }
      // Stop polling and initialize checkout when polling elements are not found
      clearInterval(clear);
      updateGlobalOpts();
      initCheckout();
      return;
    }, 100);
  }

  function updateGlobalOpts() {
    var $orderTotal = document.querySelector(ORDER_TOTAL_SELECTOR);
    var newCustomTotal = $orderTotal ? parseInt($orderTotal.getAttribute(ORDER_TOTAL_DATA_ATTRIBUTE)) : null;
    if (!newCustomTotal || isNaN(newCustomTotal) || newCustomTotal === 0) {
      console.warn(ERROR_LABEL + 'failed to query updated order total');
      return;
    }
    window.BreadShopifyPlus.opts.customTotal = newCustomTotal;

    var $taxTotal = document.querySelector(TAX_TOTAL_SELECTOR);
    var newTaxTotal = $taxTotal ? parseInt($taxTotal.getAttribute(TAX_TOTAL_DATA_ATTRIBUTE)) : null;
    if (newTaxTotal === null || isNaN(newTaxTotal)) {
      console.warn(ERROR_LABEL + 'failed to query updated tax amount');
      return;
    }
    window.BreadShopifyPlus.opts.tax = newTaxTotal;
  }

  function initCheckout() {
    if (window.BreadShopifyPlus === undefined) {
      console.warn('[Bread-Shopify] BreadShopifyPlus object undefined, exiting');
      return;
    }

    PLACE_ORDER_BTN_SELECTOR = selectors.placeOrderBtn || GET_PLACE_ORDER_SELECTOR(); // Update place order button selector now that DOM is loaded
    var billingInputNode = document.querySelector(BILLING_INPUT_SELECTOR);
    var placeOrder = document.querySelector(PLACE_ORDER_BTN_SELECTOR);
    if (!placeOrder || !billingInputNode) {
      console.warn('[Bread-Shopify] Place order button or billing input not detected, exiting');
      return;
    }

    // Find and process Bread gateways
    var breadGateways = [];
    Array.prototype.forEach.call(document.querySelectorAll(ALL_GATEWAY_SELECTOR), function(g) {
      if (g.getAttribute('data-gateway-group') !== 'offsite') {
        // Make sure non-"offsite" gateways show "Place order" button and billing address form
        // Consider changing this event handler to 'change'
        g.querySelector('input').addEventListener('click', handleOtherPaymentMethodClick);
      } else {
        var gatewayName = g.querySelector('label').innerText.trim();
        if (BREAD_GATEWAY_KEY[gatewayName]) {
          var breadGateway = {
            // labelNode: g.querySelector('label'),
            input: g.querySelector('input'),
            checkoutNode: document.querySelector(CHECKOUT_NODE_SELECTOR(g)),
            checkoutNodeId: CHECKOUT_NODE_ID_CREATOR(g)
          };
          breadGateways.push(breadGateway);
          if (breadGateway.input.checked) {
            showPlaceOrderButton(false);
            showShopifyBillingAddressForm(false);
          }
          processBreadGateway(breadGateway);
        } else {
          // Make sure non-Bread "offsite" gateways show "Place order" button and billing address form
          g.querySelector('input').addEventListener('click', handleOtherPaymentMethodClick);
        }
      }
    });

    if (breadGateways.length < 1) {
      console.warn('[Bread-Shopify] No Bread gateways detected, exiting');
      return;
    }

    runAfterLDFlagsLoad(overwriteBreadLogo);
  }

  function processBreadGateway(g) {
    // Clear out contents for embedded checkout node
    clearNodeContents(g.checkoutNode);

    // Add padding, min-height, and background color
    formatCheckoutNode(g.checkoutNode, BREAD_FORM_ID);

    validateOptsItems();

    // Run initial checkout
    bread.checkout(Opts());
    
    // Add change event handler for Bread gateway
    // g.labelNode.addEventListener('click', handleBreadLabelClick);
    g.input.addEventListener('change', handleBreadInputChange.bind(g));
  }

  function runAfterLDFlagsLoad() {
    var MAX_SECS_BEFORE_ABORT = 5;
    var TIMEOUT_INTERVAL = 100;
    var RETRIES = MAX_SECS_BEFORE_ABORT * 1000 / TIMEOUT_INTERVAL;
    var ERROR_MSG = 'Error polling LD flags: ';
    var TIMEOUT_ERROR = 'Could not read LD flags within ' + MAX_SECS_BEFORE_ABORT + ' seconds.';

    var args = Array.prototype.slice.call(arguments);
    if (args.length < 1) {
      console.warn(ERROR_LABEL + ERROR_MSG + 'no arguments to run');
      return;
    }

    var retryCount = 0;
    var retry = setInterval(function() {
      try {
        if (window.bread && window.bread.ldflags && window.bread.ldflags._isReady) {
          window.clearInterval(retry);
          args.forEach(function(fn) {
            fn();
          });
        }
        if (retryCount < RETRIES) {
          retryCount += 1;
        } else {
          throw new Error(TIMEOUT_ERROR);
        }
      } catch (err) {
        window.clearInterval(retry);
        console.warn(ERROR_LABEL + ERROR_MSG + err.message);
      }
    }, TIMEOUT_INTERVAL);
  }

  function overwriteBreadLogo() {
    if (window.bread && window.bread.ldflags['multipay-enable'] === false) {
      return;
    }

    var styles = document.createElement('style');
    styles.innerHTML = '#bread-logo-label{ font-family: "Helvetica Neue"; }#bread-logo-installments{ color: #5156ea; }#bread-logo-installments:hover{ filter: brightness(1.1); }#bread-logo-splitpay{ color: #57c594; }#bread-logo-splitpay:hover{ filter: brightness(1.1); }';;
    document.body.appendChild(styles);
    var img = document.querySelector('img[alt="Bread"]');
    if (img && img.parentElement) img.parentElement.innerHTML = '<span id="bread-logo-label">Pay over time with <span id="bread-logo-installments">Installments</span> or <span id="bread-logo-splitpay">SplitPay</span></span>';
  }

  // Opts generator
  function Opts(opts) {
    if (!opts) opts = {};

    var newOpts = {
      formId: opts.formId ? opts.formId : BREAD_FORM_ID,
      financingProgramId: getFinancingProgramID(),
      actAsLabel: false,
      asLowAs: true,
      buttonLocation: 'checkout',
      disableEditShipping: true,
      hideFieldsWhenProvided: false,
      billingContact: opts.billingContact ? opts.billingContact : window.BreadShopifyPlus.opts.billingContact,
      shippingContact: opts.shippingContact ? opts.shippingContact : window.BreadShopifyPlus.opts.shippingContact,
      items: window.BreadShopifyPlus.opts.items,
      discounts: window.BreadShopifyPlus.opts.discounts.filter(function(d) { return d.amount > 0 }),
      shippingOptions: window.BreadShopifyPlus.opts.shippingOptions,
      customTotal: window.BreadShopifyPlus.opts.customTotal,
      done: function(err, tx) {
        if (err) {
          logError(err, "Error returned to done callback", tx);
          return;
        }

        AJAX({
          type: 'POST',
          url: TX_RECORD_URL,
          contentType: 'application/json; charset=utf-8',
          data: {
            checkoutID: window.BreadShopifyPlus.reference,
            transactionID: tx
          },
          callback: function(data, err) {
            if (err) {
              logError(err, "Executing done callback failed", tx)
            } else {
              // Place order button
              try {
                document.querySelector(PLACE_ORDER_BTN_SELECTOR).click();
              } catch(err) {
                logError(err, "Executing done callback failed", tx)
              }
            }
          }
        });
      },
      calculateTax: function(sc, bc, cb) {
        // Use calculateTax callback to update Shopify billing address form
        if (billingShippingSame(bc, sc)) {
          document.querySelector(SHIPPING_INPUT_SELECTOR).click();
        } else {
          document.querySelector(BILLING_INPUT_SELECTOR).click();
          enterBillingAddressForShopify(bc);
        }
        // Always resolve this static Tax amount
        cb(null, window.BreadShopifyPlus.opts.tax);
      }
    };

    if (HEALTH_CARE) {
      ['items', 'discounts', 'shippingOptions'].forEach(function(e) {  
        delete newOpts[e];
      });
    }

    return newOpts;
  }

  function isValidUUID(input) {
    return RegExp(/^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/).test(input);
  }

  function getFinancingProgramID() {
    var financingProgramID = '';
    if (TARGETED_FINANCING && isValidUUID(window.BreadShopifyPlus.opts.financingProgramId)) {
      // Use targeted financing program provided by custom implementation
      financingProgramID = window.BreadShopifyPlus.opts.financingProgramId;
    } else if (TARGETED_FINANCING && window.BreadShopifyPlus.opts.customTotal >= TARGETED_FINANCING_THRESHOLD) {
      // Use targeted financing program determined by gateway account settings
      financingProgramID = TARGETED_FINANCING_PROGRAM_ID;
    }
    return financingProgramID;
  }

  // Event handlers
  function handleBreadLabelClick(e) {
    e.stopPropagation();
    showPlaceOrderButton(false);
    showShopifyBillingAddressForm(false);
  }

  function handleBreadInputChange(e) {
    if (!this.input.checked) {
      return;
    }
    showPlaceOrderButton(false);
    showShopifyBillingAddressForm(false);
    // Use Shopify's billing address if it is valid and different from shipping
    if (document.querySelector(BILLING_INPUT_SELECTOR).checked && validateBillingAddressNodes(getBillingAddressNodes())) {
      BILLING_INPUT_CHANGE++;
      bread.checkout(Opts({ billingContact: getBillingContactFromShopify() }));
    } else if (BILLING_INPUT_CHANGE > 0) {
      bread.checkout(Opts());
    }
  }

  function handleOtherPaymentMethodClick(e) {
    showPlaceOrderButton(true);
    showShopifyBillingAddressForm(true);
  }

  // DOM manipulators
  function showPlaceOrderButton(show) {
    if (typeof show !== 'boolean') {
      return;
    }
    var style = show ? 'block' : 'none';
    document.querySelector(PLACE_ORDER_BTN_SELECTOR).style.display = style;
  }

  function showShopifyBillingAddressForm(show) {
    if (typeof show !== 'boolean') {
      return;
    }
    var style = show ? 'block' : 'none';
    document.querySelector(BILLING_ADDRESS_FORM_SELECTOR).style.display = style;
  }

  // Billing address utility functions
  function billingShippingSame(b, s) {
    return b.firstName === s.firstName &&
      b.lastName === s.lastName &&
      b.address === s.address &&
      b.address2 === s.address2 &&
      b.city === s.city &&
      b.state === s.state &&
      b.zip === s.zip &&
      b.phone.replace(/\D/g,'') === s.phone.replace(/\D/g,'');
  }

  function getBillingAddressNodes() {
    return {
      firstName: document.querySelector(BILLING_ADDRESS_FIRST_NAME),
      lastName: document.querySelector(BILLING_ADDRESS_LAST_NAME),
      address: document.querySelector(BILLING_ADDRESS_ADDRESS_1),
      address2: document.querySelector(BILLING_ADDRESS_ADDRESS_2),
      city: document.querySelector(BILLING_ADDRESS_CITY),
      country: document.querySelector(BILLING_ADDRESS_COUNTRY),
      state: document.querySelector(BILLING_ADDRESS_PROVINCE),
      zip: document.querySelector(BILLING_ADDRESS_ZIP),
      phone: document.querySelector(BILLING_ADDRESS_PHONE)
    };
  }

  // Get the billing address state formatted as state code
  // Shopify's previous billing state <select> field expects full state name values
  // Shopify's new billing state <select> field expects state code values
  function getBillingStateCode($state) {
    var state = $state.querySelector('select').value;
    return state.length === 2 ? state : $state.querySelector('option[value="' + state + '"]').getAttribute('data-code');
  }

  function getBillingAddressValues($address) {
    return {
      firstName: $address.firstName.querySelector('input').value,
      lastName: $address.lastName.querySelector('input').value,
      address: $address.address.querySelector('input').value,
      address2: $address.address2.querySelector('input').value,
      city: $address.city.querySelector('input').value,
      state: getBillingStateCode($address.state),
      zip: $address.zip.querySelector('input').value,
      phone: $address.phone ? $address.phone.querySelector('input').value : null
    }
  }

  function validateBillingAddressNodes(nodes) {
    var valid = true;
    for(k in nodes) {
      if (k !== "address2"){
        var i = nodes[k].querySelector('input');
        if (!i) {
          i = nodes[k].querySelector('select');
        }
        if (i.value === '') {
          valid = false;
          // nodes[k].classList.add('field--error');
        }
      }
    }
    return valid;
  }

  function clearBillingAddressErrors() {
    var nodes = getBillingAddressNodes();
    for(k in nodes) {
      nodes[k].classList.remove('field--error');
    }
  }

  function getBillingContactFromShopify() {
    var shopifyContact = getBillingAddressValues(getBillingAddressNodes());
    shopifyContact.email = window.BreadShopifyPlus.opts.shippingContact.email;
    if (shopifyContact.phone === null) shopifyContact.phone = window.BreadShopifyPlus.opts.shippingContact.phone;
    return shopifyContact;
  }

  function enterBillingAddressForShopify(bc) {
    $address = getBillingAddressNodes();
    $address.firstName.querySelector('input').value = bc.firstName;
    $address.lastName.querySelector('input').value = bc.lastName;
    $address.address.querySelector('input').value = bc.address;
    $address.address2.querySelector('input').value = bc.address2;
    $address.city.querySelector('input').value = bc.city;
    $address.zip.querySelector('input').value = bc.zip;

    // Format state code to state name
    $address.state.querySelector('select').value = BILLING_STATE_OPTION_VALUE($address.state, bc.state);
    if ($address.country) $address.country.querySelector('select').value = 'United States';
    if ($address.phone) $address.phone.querySelector('input').value = bc.phone;
  }

  function loadCheckoutJS(breadScriptSrc, breadAPIKey) {
    var init = false;
    function scriptReady() {
      if (!init && (!document.readyState || document.readyState === 'loaded' || document.readyState === 'interactive' || document.readyState === 'complete')) {
        init = true;
        pollCheckoutElement();
      }
    }

    function pageReady() {
      if (!init && 'bread' in window) {
        init = true;
        pollCheckoutElement();
      }
    }
    var script = document.createElement('script');
    script.type = 'text/javascript';
    script.src = breadScriptSrc;
    script.setAttribute('data-api-key', breadAPIKey);
    script.onload = script.onreadystatechange = scriptReady;
    document.head.appendChild(script);
    document.addEventListener('DOMContentLoaded', pageReady);
    document.onreadystatechange = pageReady();
  }

  function clearNodeContents(n) {
    while(n.firstChild) {
      n.removeChild(n.firstChild);
    }
  }

  // Format gateway node for embedded checkout and return the opts.formId
  function formatCheckoutNode(n, formId) {
    if (typeof formId === "undefined") formId = BREAD_FORM_ID;
    n.style.padding = '0';
    n.style.background = '#fff';
    n.style.minHeight = '401px';

    var fragment = document.createDocumentFragment();
    if (HEALTH_CARE) {
      fragment.appendChild(getHealthCareDisclosure());
    }
    var breadCheckoutForm = document.createElement('div');
    breadCheckoutForm.setAttribute('id', formId);
    fragment.appendChild(breadCheckoutForm);
    n.appendChild(fragment);
  }

  function getHealthCareDisclosure() {
    var d = document.createElement('div');
    d.setAttribute('id', 'bread-healthcare-disclosure');
    var p = document.createElement('p');
    p.innerHTML = HEALTH_CARE_DISCLOSURE;
    p.setAttribute('id', 'bread-healthcare-content');
    p.style.paddingBottom = '10px';
    d.appendChild(p);
    return d;
  }

  function AJAX(o) {
    var r = new XMLHttpRequest();
    r.open(o.type, o.url, true);
    if (o.type === "POST") {
      r.setRequestHeader('Content-Type', o.contentType || 'application/json; charset=utf-8');
    }

    r.onload = function() {
      if (r.status >= 200 && r.status < 400) {
        o.callback(JSON.parse(r.responseText), null);
      } else {
        o.callback(null, r.responseText);
      }
    };

    r.onerror = function() {
      o.callback(null, "(AJAX) unable to reach " + o.url);
    };

    o.data ? r.send(JSON.stringify(o.data)) : r.send();
  };

  function logError(err, message, transactionID) {
    AJAX({
      type: "POST",
      url: "/apps/bread/errors",
      contentType: "application/json; charset=utf-8",
      data: {
        error: err.toString(),
        message: message,
        stackTrace: err.stack,
        userAgent: window.navigator.userAgent,
        referrer: window.location.href,
        pageType: "checkout",
        apiKey: BREAD_API_KEY,
        gatewayReference: window.BreadShopifyPlus.reference,
        transactionID: transactionID,
      },
      callback: function(data, err) {
        return;
      }
    });
  }

})(window, document);

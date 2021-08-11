(function() {
  var activeVersion = "{{ .ActiveVersion }}"
  if (activeVersion !== "classic") {
    return
  }
  
  var ranCreateButtons = false;
  var SPLITPAY_PROMO_PRICE_SELECTOR = '.bread-splitpay-price';
  var SPLITPAY_PROMO_BTN_SELECTOR = '.bread-splitpay-btn';

  // Inject BreadJS onto the page
  var script = document.createElement('script');
  script.type = 'text/javascript';
  script.src = "{{ .BreadJS }}";
  script.setAttribute("data-api-key", "{{ .ApiKey }}");
  script.onload = script.onreadystatechange = breadScriptReady;
  (document.getElementsByTagName("head")[0] || document.documentElement).appendChild(script);

  document.addEventListener('DOMContentLoaded', function(){ breadPageReady(); });
  document.onreadystatechange = breadPageReady();

  // Templated settings
  var asLowAs = false;
  {{ if .AsLowAs }}
  asLowAs = true;
  {{ end }}

  var customCSS = undefined;
  {{ if .CSS }}
  customCSS = '{{ .CSS }}';
  {{ end }}

  var customCSSCart = customCSS;
  {{ if .CSSCart }}
  customCSSCart = '{{ .CSSCart }}';
  {{ end }}

  var actAsLabel = false;
  {{ if .ActAsLabel }}
  actAsLabel = true; {{end}}

  var allowCheckoutPDP = true;
  {{ if not .AllowCheckoutPDP }}
  allowCheckoutPDP = false;
  {{ end }}

  var enableAddToCart = true;
  {{ if not .EnableAddToCart }}
  enableAddToCart = false;
  {{ end }}

  var allowCheckoutCart = true;
  {{ if not .AllowCheckoutCart }}
  allowCheckoutCart = false;
  {{ end }}

  var healthcareMode = false;
  {{ if .HealthcareMode }}
  healthcareMode = true;
  {{ end }}

  var targetedFinancing = false;
  {{ if .TargetedFinancing }}
  targetedFinancing = true;
  {{ end }}

  var targetedFinancingID = '';
  {{ if .TargetedFinancingID }}
  targetedFinancingID = '{{ .TargetedFinancingID }}';
  {{ end }}

  var targetedFinancingThreshold = 0;
  {{ if .TargetedFinancingThreshold }}
  targetedFinancingThreshold = parseInt(Math.round({{ .TargetedFinancingThreshold }} * 100));
  {{ end }}

  var calculateTaxWithDrafts = false;
  {{ if .CalculateTaxDraftOrder }}
  calculateTaxWithDrafts = true;
  {{ end }}

  function breadScriptReady() {
    if (!ranCreateButtons && (!document.readyState || document.readyState === "loaded" || document.readyState === "interactive" || document.readyState === "complete")) {
      ranCreateButtons = true;
      createButtons();
    }
  }

  function breadPageReady() {
    if (!ranCreateButtons && "bread" in window) {
      ranCreateButtons = true;
      createButtons();
    }
  }

  /**
   * Public window API
   */

  var optsByButtonId = {};
  var willCheckoutWithOpts = undefined;

  var BreadShopify = {
    // The provided callback should accept the generic opts for a button and callback.
    // The callback needs to be invoked with the modified options in order to enable the button.
    setWillCheckoutWithOpts: function(f) {
      // Asset function passed in
      if (typeof f !== "function") {
        console.warn(ERROR_LABEL + 'BreadShopify.setWillCheckoutWithOpts requires a function as the sole argument');
        return;
      }

      willCheckoutWithOpts = f;
    },
    optsForButtonId: function(id) {
      if (typeof optsByButtonId[id] !== "undefined") {
        return optsByButtonId[id];
      }
      return {};
    }
  };
  window.BreadShopify = BreadShopify;

  /**
   * Internal functions
   */

  var ERROR_LABEL = '[Bread-Shopify] ';
  BreadError = new BreadErrorController();

  function checkoutWithOpts(opts) {
    if (typeof willCheckoutWithOpts === "function") {
      willCheckoutWithOpts(opts, function(err, newOpts) {
        if (err != undefined) {
          console.log(ERROR_LABEL + 'Not adding behavior to Bread button [%s]', newOpts.buttonId);
          return;
        }
        optsByButtonId[opts.buttonId] = newOpts;
        bread.checkout(newOpts);
      });
    } else {
      // Checkout
      optsByButtonId[opts.buttonId] = opts;
      bread.checkout(opts);
    }
  };

  function configureCartButton() {
    getCart(function(err, cart) {
      if (err) {
        return console.log(ERROR_LABEL + 'query for shopify cart produced err: ', err);
      }

      // Determine if cart has products that require shipping
      var requiresShipping = false;
      var count = 0;
      while (!requiresShipping && count < cart.items.length) {
        requiresShipping = cart.items[count].requires_shipping;
        count++;
      }

      // Generate opts
      var opts = {};
      opts.buttonId = 'bread-checkout-btn';
      opts.buttonLocation = 'cart_summary';
      opts.actAsLabel = false;
      opts.items = cart.items.map(mapShopifyItemToBreadItem);
      opts.onCustomerOpen = onCustomerOpen;
      opts.onCustomerClose = onCustomerClose;
      opts.done = function(err, token) {
        AJAX({
          "type": "POST",
          "url": "/apps/bread/orders",
          "contentType": "application/json; charset=utf-8",
          "data": {
            "transactionId": token
          },
          "callback": function(data, err) {
            if (err) {
              var errorType = err === 'remainderPay' ? err : 'transaction';
              BreadError.throw(errorType, err);
            } else {
              clearCart(redirectCustomConfirmation.bind(null, data.orderId));
            }
          }
        });
      };
      opts.calculateTax = function(shippingContact, callback) {
        var lineItems = cart.items.map(mapShopifyItemToLineItem);
        calculateTax(shippingContact, lineItems, callback);
      };
      if (requiresShipping) {
        opts.calculateShipping = getCartShippingRates;
      } else {
        opts.shippingOptions = getShippingNotRequiredOption();
      }
      opts.asLowAs = asLowAs;
      opts.customCSS = customCSSCart;
      opts.allowCheckout = allowCheckoutCart && validateItemSkus(opts.items);
      opts.allowSplitPayCheckout = false;
      if (healthcareMode) {
        opts.customTotal = opts.items.reduce(function(total, item) {
          return total + (item.price * item.quantity);
        }, 0);
        opts.items = [];
        opts.allowCheckout = false;
      }
      if (targetedFinancing && cart.total_price >= targetedFinancingThreshold) {
        opts.financingProgramId = targetedFinancingID;
      }
      checkoutWithOpts(opts);
      setupSplitPayPromos(opts);
    });
  };

  function configureProductButton() {
    var handleSetup = function(product, variant) {
      var items = [mapShopifyVariantToBreadItem(product, variant)];
      var lineItems = [mapShopifyVariantToLineItem(product, variant)];
      setProductButton(items, lineItems, variant.requires_shipping, variant.available);
    };
    queryProductByHandleInUrl(function(product, err) {
      if (err) {
        return console.log(ERROR_LABEL + 'Query for product returned an error' + err);
      }

      // Bootstrap button with first variant
      var vid = product.variants[0].id;
      handleSetup(product, product.variants[0]);
      if (product.variants.length <= 1) return;

      var variantIntervalId = setInterval(function() {
        var newVidEl = document.querySelectorAll("select[name='id'] :checked");
        if (newVidEl.length > 0) {
          var newVid = newVidEl[0].getAttribute('value');
          if (vid !== newVid) {
            vid = newVid;
            var variant = product.variants[variantIndexFromVid(product, vid)];
            handleSetup(product, variant);
          }
        }
      }, 500);
    });
  };

  function setProductButton(items, lineItems, requiresShipping, available) {
    // Generate opts
    var opts = {};
    opts.buttonId = 'bread-checkout-btn-product';
    opts.buttonLocation = 'product';
    opts.items = items;
    opts.onCustomerOpen = onCustomerOpen;
    opts.onCustomerClose = onCustomerClose;
    opts.calculateTax = function(shippingContact, cb) {
      calculateTax(shippingContact, lineItems, cb);
    };

    if (requiresShipping) {
      opts.calculateShipping = function(shippingContact, cb) {
        var payload = {
          "shopName": Shopify.shop.split(".")[0],
          "state": shippingContact.state,
          "zip": shippingContact.zip,
          "lineItems": lineItems
        };
        AJAX({
          "type": "POST",
          "url": "/apps/bread/cart/shipping",
          "contentType": "application/json; charset=utf-8",
          "data": payload,
          "callback": function(rates, err) {
            if (err) {
              // Correctly handle products that do not require shipping
              if (err === '{"error":["This cart does not require shipping"]}') {
                var noShip = getShippingNotRequiredOption();
                cb(null, noShip);
              } else {
                BreadError.throw('shipping', err);
              }
            } else if (!rates || rates.length === 0) {
              BreadError.throw('shipping', 'No shipping rates found');
            } else {
              cb(null, convertRatesToShippingOptions(rates));
            }
          }
        });
      };
    } else {
      opts.shippingOptions = getShippingNotRequiredOption();
    }

    if (enableAddToCart && lineItems.length === 1 && available && !healthcareMode) {
      opts.addToCart = function(err, cb) {
        AJAX({
          "type": "POST",
          "url": "/cart/add.js",
          "contentType": "application/json; charset=utf-8",
          "data": {
            "id": lineItems[0].id,
            "quantity": lineItems[0].quantity
          },
          "callback": function(data, err) {
            if (err) {
              err = JSON.parse(err);
              cb(err.description);
              return
            }
            setTimeout(function() {
              window.location = "/cart";
            }, 1000);
            cb(null);
          }
        });
      };
    }

    opts.done = function(err, token) {
      AJAX({
        "type": "POST",
        "url": "/apps/bread/orders",
        "contentType": "application/json; charset=utf-8",
        "data": {
          "transactionId": token
        },
        "callback": function(data, err) {
          if (err) {
            var errorType = err === 'remainderPay' ? err : 'transaction';
            BreadError.throw(errorType, err);
          } else {
            clearCart(redirectCustomConfirmation.bind(null, data.orderId));
          }
        }
      });
    };
    opts.actAsLabel = actAsLabel;
    opts.asLowAs = asLowAs;
    opts.customCSS = customCSS;
    opts.allowCheckout = allowCheckoutPDP && available && validateItemSkus(opts.items);
    opts.allowSplitPayCheckout = false;
    if (healthcareMode) {
      opts.customTotal = opts.items[0].price * opts.items[0].quantity;
      opts.items = [];
      opts.allowCheckout = false;
    }
    if (targetedFinancing) {
      var cartTotal = opts.items.reduce(function(total, item) {
        return total + (item.price * item.quantity);
      }, 0);
      if (cartTotal >= targetedFinancingThreshold) {
        opts.financingProgramId = targetedFinancingID;
      }
    }
    checkoutWithOpts(opts);
    setupSplitPayPromos(opts);

    //Watch quantity change
    var q = document.querySelector('input[name="quantity"]');
    if (q != null) {
      q.addEventListener('input', function(e){
        var itemQuantity = getQuantityValue();
        if (!isNaN(itemQuantity) && itemQuantity != 0) {
          var newOpts = optsByButtonId['bread-checkout-btn-product'];

          newOpts.items[0].quantity = itemQuantity;
          checkoutWithOpts(newOpts);
        }
      });
    }
  };

  function setupSplitPayPromos(opts) {
    var showSplitPayPrice = document.querySelector(SPLITPAY_PROMO_PRICE_SELECTOR) !== null;
    var showSplitPayBtn = document.querySelector(SPLITPAY_PROMO_BTN_SELECTOR) !== null;

    if (!showSplitPayPrice && !showSplitPayBtn) {
      return;
    }

    var MAX_SECS_BEFORE_ABORT = 5;
    var TIMEOUT_INTERVAL = 100;
    var RETRIES = MAX_SECS_BEFORE_ABORT * 1000 / TIMEOUT_INTERVAL;
    var ERROR_MSG = 'Could not setup promotional label for SplitPay: ';
    var INTEGRATION_ERROR_MSG = 'Could not create Bread SplitPay Promotional Label within ' + MAX_SECS_BEFORE_ABORT + ' seconds.';

    var retryCount = 0;
    var retry = setInterval(function() {
      try {
        if (window.bread && window.bread.ldflags && window.bread.ldflags._isReady) {
          window.clearInterval(retry);
          if (window.bread.ldflags['multipay-enable']) {
            showSplitPayPromos(opts, showSplitPayPrice, showSplitPayBtn);
          }
        }
        if (retryCount < RETRIES) {
          retryCount += 1;
        } else {
          throw new Error(ERROR_LABEL + INTEGRATION_ERROR_MSG);
        }
      } catch (err) {
        console.warn(ERROR_LABEL + ERROR_MSG + err);
        window.clearInterval(retry);
      }
    }, TIMEOUT_INTERVAL);
  };

  function showSplitPayPromos(opts, showSplitPayPrice, showSplitPayBtn) {
    var total = null;
    if (opts.hasOwnProperty('customTotal')) {
      total = opts.customTotal;
    } else if (opts.hasOwnProperty('items')) {
      total = opts.items.reduce(function(sum, i) {
        return (i.price * i.quantity) + sum;
      }, 0);
    } else {
      console.warn(ERROR_LABEL + 'failed to calclulate total for SplitPay promos.');
    }

    opts.allowSplitPayCheckout = false;

    if (!total || total > 100000) {
      return;
    }

    if (showSplitPayPrice) {
      document.querySelector(SPLITPAY_PROMO_PRICE_SELECTOR).style.display = 'block';
      bread.showSplitPayPromo({
        selector: SPLITPAY_PROMO_PRICE_SELECTOR,
        total: total,
        includeInstallments: true,
        openModalOnClick: true,
        opts: opts
      });
    }

    if (showSplitPayBtn) {
      document.querySelector(SPLITPAY_PROMO_BTN_SELECTOR).style.display = 'block';
      bread.showSplitPayPromo({
        selector: SPLITPAY_PROMO_BTN_SELECTOR,
        total: total,
        includeInstallments: false,
        openModalOnClick: true,
        opts: opts
      });
    }
  }

  function variantIndexFromVid(product, vid) {
    for (var x = 0; x < product.variants.length; x++) {
      if (product.variants[x].id == vid) return x;
    }
    return 0;
  };

  function redirectToCart(msg) {
    if (!msg) {
      msg = "Bread was unable to authorize your transaction. Please checkout using another payment method.";
    }
    alert(msg);
    window.location = "/cart";
  };

  function getQuantityValue() {
    // Default to quantity of 1
    var v = 1;
    var q = document.querySelector('input[name="quantity"]');

    if (q != null) {
      v = parseInt(q.value, 10);
    }
    return v;
  }

  function mapShopifyVariantToBreadItem(product, variant) {
    var firstProductImage = product.images.length > 0 ? product.images[0] : "";
    return {
      "name": [product.title, variant.title].join(", "),
      "price": variant.price,
      "sku": product.id + ";::;" + variant.sku,
      "imageUrl": variant.featuredImage ? sanitizeImageUrl(variant.featuredImage) : sanitizeImageUrl(firstProductImage),
      "detailUrl": sanitizeDetailUrl(product.url),
      "quantity": getQuantityValue()
    };
  };

  function mapShopifyVariantToLineItem(product, variant) {
    return {
      "id": variant.id,
      "quantity": getQuantityValue()
    };
  };

  function mapLineItemToVariantItem(item) {
    return {
      variant_id: item.id,
      quantity: item.quantity
    };
  };

  function mapShopifyItemToBreadItem(item) {
    return {
      "name": item.variant_options.length > 1 ? item.product_title + item.variant_title : item.product_title,
      "price": item.price,
      "sku": item.product_id + ";::;" + item.sku,
      "imageUrl": sanitizeImageUrl(item.image),
      "detailUrl": sanitizeDetailUrl(item.url),
      "quantity": item.quantity
    };
  };

  function sanitizeImageUrl(url) {
    if (url === undefined || typeof url !== "string" || url.length === 0) {
      return "";
    }

    // Ensure a protocol exists, default to https
    var compUrl = url.toLowerCase();
    if (compUrl.indexOf("http") !== 0) {
      if (compUrl.indexOf("://") == 0) {
        return "http" + url;
      }
      if (compUrl.indexOf("//") == 0) {
        return "http:" + url;
      }
    }
    return url;
  };

  function sanitizeDetailUrl(url) {
    if (url === undefined || typeof url !== "string" || url.length === 0) {
      return "";
    }

    // Ensure the url includes the full path, add if necessary
    var compUrl = url.toLowerCase();
    if (compUrl.indexOf("/products") == 0) {
      return window.location.protocol + "//" + window.location.hostname + url;
    }
    return url;
  };

  function mapShopifyItemToLineItem(item) {
    return {
      "id": item.id,
      "quantity": item.quantity
    };
  };

  function convertRatesToShippingOptions(rates) {
    var shippingOpts = [];
    var i = 0;
    for (i = 0; i < rates.length; i++) {
      var rate = rates[i];
      shippingOpts.push({
        cost: Math.round(parseFloat(rate.price) * 100),
        type: rate.name,
        typeId: rate.code
      });
    }
    return shippingOpts;
  };

  function calculateTax(shippingContact, lineItems, callback) {
    var url = '/apps/bread/cart/tax';
    if (calculateTaxWithDrafts) {
      // Use alternate calculateTax endpoint and format lineItems for draft orders
      url = '/apps/bread/cart/tax/draftorder';
      lineItems = lineItems.map(mapLineItemToVariantItem);
    }
    if (!shippingContact.selectedShippingOption) {
      BreadError.throw('tax', 'No shipping rates found');
      return;
    }
    var payload = {
      shopName: Shopify.shop.split(".")[0],
      state: shippingContact.state,
      zip: shippingContact.zip,
      lineItems: lineItems,
      shippingAddress: shippingContact,
      shippingLine: {
        handle: null,
        title: shippingContact.selectedShippingOption.type,
        price: shippingContact.selectedShippingOption.cost / 100
      }
    };
    AJAX({
      "type": "POST",
      "url": url,
      "contentType": "application/json; charset=utf-8",
      "data": payload,
      "callback": function(data, err) {
        if (err) {
          callback(err);
        } else {
          callback(null, parseInt(data.totalTax));
        }
      }
    });
  };

  function onCustomerOpen(err, data, cb) {
    BreadError.close();
    cb(data);
  }

  function onCustomerClose(err, data) {
    BreadError.close();
    return;
  }

  function convertCartToMiniLineItems(cc) {
    var li = [];
    cc.items.forEach(function(i) {
      li.push({
        "id": i.id,
        "quantity": i.quantity
      });
    });
    return li;
  };

  function encodeQueryData(data) {
    var ret = [];
    for (var d in data) {
      ret.push(encodeURIComponent(d) + "=" + encodeURIComponent(data[d]));
    }
    return ret.join("&");
  }

  // Will get the shipping rates for the items on the cart
  function getCartShippingRates(shippingContact, cb) {
    var queryData = {
      "shipping_address[zip]": shippingContact.zip,
      "shipping_address[country]": "United States", // Hard coded
      "shipping_address[province]": shippingContact.state
    };
    var query = encodeQueryData(queryData);
    var url = "/cart/shipping_rates.json?" + query;

    AJAX({
      "type": "GET",
      "url": url,
      "callback": function(shippingOptions, err) {
        if (err) {
          // Correctly handle products that do not require shipping
          if (err === '{"error":["This cart does not require shipping"]}') {
            var noShip = getShippingNotRequiredOption();
            cb(null, noShip);
          } else {
            BreadError.throw('shipping', err);
          }
        } else if (!shippingOptions.shipping_rates || shippingOptions.shipping_rates.length === 0) {
          BreadError.throw('shipping', 'No shipping rates found');
        } else {
          var rates = convertRatesToShippingOptions(shippingOptions.shipping_rates);
          cb(null, rates);
        }
      }
    });
  };

  function getShippingNotRequiredOption() {
    return [{ cost: 0, type: 'Shipping not required', typeId: 'shipping-not-required'}];
  }

  function clearCart(cb) {
    AJAX({
      "type": "POST",
      "url": "/cart/clear.js",
      "contentType": "application/json; charset=utf-8",
      "data": {},
      "callback": cb
    });
  };

  function redirectCustomConfirmation(orderId) {
    window.location = "/apps/bread/orders/confirmation/" + orderId;
  };

  function getCart(cb) {
    AJAX({
      "url": "/cart.js",
      "type": "GET",
      "callback": function(cart, err) {
        if (err) {
          cb(err);
        } else {
          cb(null, cart);
        }
      }
    });
  };

  function queryProductByHandleInUrl(cb) {
    var pathName = window.location.pathname;

    //Remove trailing slash from pathname
    pathName = pathName.slice(-1) === "/" ? pathName.slice(0,-1) : pathName;

    var pieces = pathName.split("/");

    var productName = pieces[pieces.length - 1];
    AJAX({
      "type": "GET",
      "url": "/products/" + productName + ".js",
      "callback": cb
    });
  };

  function validateItemSkus(items) {
    if (!Array.isArray(items)) return false;
    var valid = true;
    items.forEach(function(i) {
      var pieces = i.sku.split(';::;');
      if (pieces[1] === '') {
        console.warn(ERROR_LABEL + 'Checkout disabled because Product ID ' + pieces[0] + ' is missing a SKU. Please add unique SKUs to each product and variant in Shopify.');
        valid = false;
      }
    });
    return valid;
  }

  function createButtons() {
    if (document.querySelectorAll("#bread-checkout-btn").length > 0) {
      configureCartButton();
    }

    if (document.querySelectorAll("#bread-checkout-btn-product").length > 0) {
      configureProductButton();
    }
  }

  function BreadErrorController() {
    var controller = {};
    var startPositionBottom = '-50px';
    var endPositionBottom = '10px';
    var startPositionTop = '30px';
    var endPositionTop = '90px';

    var createContainer = function() {
      var c = document.createElement('div');
      c.style.setProperty('display', 'none');
      c.style.setProperty('position', 'fixed');
      c.style.setProperty('width', '100%');
      c.style.setProperty('z-index', '2147483647');
      c.style.setProperty('bottom', startPositionBottom);
      c.style.setProperty('padding', '20px');
      c.style.setProperty('font-family', 'sans-serif');
      c.style.setProperty('font-size', '20px');
      c.style.setProperty('cursor', 'pointer');
      c.style.setProperty('transition', 'all 0.4s ease');
      c.style.setProperty('opacity', '0');
      return c;
    };

    var createInnerContainer = function() {
      var i = document.createElement('div');
      i.style.setProperty('height', '100%');
      i.style.setProperty('padding', '20px 20px');
      i.style.setProperty('background', '#f3645e');
      i.style.setProperty('color', '#fff');
      return i;
    };

    var createMessageElement = function() {
      var m = document.createElement('p');
      m.style.setProperty('display', 'inline-block');
      m.style.setProperty('width', '90%');
      m.style.setProperty('margin', '0');
      return m;
    };

    var createCloseElement = function() {
      var c = document.createElement('span');
      c.style.setProperty('position', 'absolute');
      c.style.setProperty('top', '27px');
      c.style.setProperty('right', '45px');
      c.style.setProperty('font-size', '1.5em');
      c.style.setProperty('cursor', 'pointer');
      c.innerHTML = '&times';
      return c;
    };

    var createErrorMessage = function(type) {
      var label = '<span style="font-weight:bold;margin-right:5px;">Error:</span> ';
      var defaultAction = ' Please close the form, add the item to your cart, and select "Pay over time" at checkout.';
      var remainderPayAction = ' Please use a different card or contact your bank. Otherwise, you can still check out with an amount covered by your Bread loan capacity.';
      var errorTypes = {
        default: 'We were unable to complete your request.',
        shipping: 'We were unable to calculate your shipping rates.',
        remainderPay: 'The credit/debit card portion of your transaction was declined.',
        tax: 'We were unable to calculate your tax.',
        transaction: 'We were unable to authorize your transaction.'
      };
      var message = errorTypes[type] ? errorTypes[type] : errorTypes['default'];
      var action = type === 'remainderPay' ? remainderPayAction : defaultAction;
      return label + message + action;
    }

    var breadModalOpen = function() {
      var modal = document.querySelector('#bread-modal');
      return modal !== null && modal.style.display === 'block';
    }

    controller.updateError = function(type) {
      if (!this.container || !this.message) {
        console.error(ERROR_LABEL + 'container not found');
        return;
      }
      this.message.innerHTML = createErrorMessage(type);
    };

    controller.clearError = function() {
      if (!this.container || !this.message) {
        console.error(ERROR_LABEL + 'container not found');
        return;
      }
      this.message.innerHTML = '';
    };

    controller.resetPosition = function(top) {
      var direction = 'top';
      var unset = 'bottom';
      var position = startPositionTop;
      var endPosition = endPositionTop

      // Position error message to the bottom if Bread modal is open and 'top' override is falsey
      if (breadModalOpen() && !top) {
        direction = 'bottom';
        unset = 'top';
        position = startPositionBottom;
        endPosition = endPositionBottom;
      }

      this.container.style.removeProperty(unset);
      this.container.style.setProperty('display', 'none');
      this.container.style.setProperty('opacity', '0');
      this.container.style.setProperty(direction, position);
      setTimeout(function() {
        this.endPosition(direction, endPosition);
      }.bind(this), 0);
    };

    controller.endPosition = function(direction, position) {
      this.container.style.setProperty('opacity', '100');
      this.container.style.setProperty(direction, position);
    };

    controller.setResponsiveSize = function() {
      if (document.body.clientWidth > 500) {
        this.container.style.setProperty('font-size', '20px');
        this.close.style.setProperty('top', '27px');
        this.close.style.setProperty('right', '45px');
      } else {
        this.container.style.setProperty('font-size', '16px');
        this.close.style.setProperty('top', '30px');
        this.close.style.setProperty('right', '40px');
      }
    };

    controller.show = function(errorType, error) {
      if (!this.container) {
        console.error(ERROR_LABEL + 'container not found');
        return;
      }

      if (!document.body) {
        setTimeout(function() {
          this.show(errorType, error);
        }.bind(this), 500);
        return;
      }

      if (document.body.contains(this.container)) {
        document.body.removeChild(this.container);
      }

      this.setResponsiveSize();

      this.resetPosition(errorType === 'transaction');

      document.body.appendChild(this.container);

      this.updateError(errorType);

      if (error && typeof error === 'string') console.error(ERROR_LABEL + error);

      this.container.style.setProperty('display', 'block');
      return;
    };

    controller.hide = function() {
      if (!this.container) {
        console.error(ERROR_LABEL + 'container not found');
        return;
      }
      if (document.body.contains(this.container)) {
        this.resetPosition();
        this.clearError();
        document.body.removeChild(this.container);
      }
      return;
    };

    controller.initialize = function() {
      this.container = createContainer();
      this.inner = createInnerContainer();
      this.message = createMessageElement();
      this.close = createCloseElement();

      this.container.addEventListener('click', this.hide.bind(this));
      this.container.addEventListener('touchstart', this.hide.bind(this));

      this.inner.appendChild(this.message);
      this.inner.appendChild(this.close);
      this.container.appendChild(this.inner);
    };

    controller.initialize();

    // Public Interface
    this.throw = controller.show.bind(controller);
    this.close = controller.hide.bind(controller);
  }

  // AJAX implementation
  function AJAX(options) {
    var request = new XMLHttpRequest();
    request.open(options.type, options.url, true);
    if (options.type === "POST") {
      request.setRequestHeader('Content-Type', options.contentType);
    }

    request.onload = function() {
      if (request.status >= 200 && request.status < 400) {
        // Success
        var response = JSON.parse(request.responseText);
        options.callback(response, null);
      } else {
        // Error
        options.callback(null, request.responseText);
      }
    };

    request.onerror = function() {
      options.callback(null, "(AJAX) unable to reach " + options.url);
    };

    if (options.data) {
      request.send(JSON.stringify(options.data));
    } else {
      request.send();
    }
  };

})();

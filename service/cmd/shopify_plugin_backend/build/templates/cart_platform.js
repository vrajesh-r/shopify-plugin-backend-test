(function() {
  var activeVersion = "{{ .ActiveVersion }}"
  if (activeVersion !== "platform") {
    return;
  }
  var integrationKey = "{{ .IntegrationKey }}"
  var sdk = "{{ .BreadJS }}"

  var script = document.createElement("script");
  script.addEventListener("load", createButtons);
  script.setAttribute("async", true);
  script.setAttribute("src", sdk);
  script.setAttribute("type", "text/javascript");
  (document.getElementsByTagName("head")[0] || document.documentElement).appendChild(script);

  function getPDPDomId() {
    var domId = "bread-checkout-btn-product";
    if (document.querySelectorAll("#placement-pdp").length > 0) {
      domId = "placement-pdp";
    }

    return domId;
  }

  function getCartDomId() {
    var domId = "bread-checkout-btn";
    if (document.querySelectorAll("#placement-cart").length > 0) {
      domId = "placement-cart";
    }
    return domId;
  }

  function onApproved(application) {
    
  }

  function onCheckout(application) {
    
  }

  function configureCartButton() {
    
    getCart(function(err, cart) {
      if (err) {
        return console.log(ERROR_LABEL + 'query for shopify cart produced err: ', err);
      }

      if (cart.item_count == 0) {
        return console.log(ERROR_LABEL + 'cart is empty');
      }

      var setup = {integrationKey: integrationKey};
      window.RBCPayPlan.setup(setup);
      window.RBCPayPlan.on('INSTALLMENT:APPLICATION_DECISIONED', onApproved);
      window.RBCPayPlan.on('INSTALLMENT:APPLICATION_CHECKOUT', onCheckout); 

      var placement = {};
      placement.allowCheckout =  false;
      placement.domID = getCartDomId();
      placement.order = {};
      var orderItems = cart.items.map(shopifyItemsToBreadItems);

      var currency = cart.currency;
      orderItems = updateBreadItemsWithCurrency(orderItems, currency);
      placement.order.items = orderItems;
      placement.order.subTotal = {value: cart.items_subtotal_price, currency: currency};
      placement.order.totalPrice = {value: cart.total_price, currency: currency};
      placement.order.currency = currency
      placement.order.totalShipping = { value: 0, currency: currency }; //Not available in shopify cart object
      placement.order.totalTax =  { value: 0, currency: currency }; //Not available in shopify cart object
      placement.order.totalDiscounts =  { value: calculateTotalDiscount(cart), currency: currency };

      window.RBCPayPlan.registerPlacements([placement]);
      window.RBCPayPlan.__internal__.init();
    });
  };

  function shopifyItemsToBreadItems(item) {
    return {
      "name": item.variant_options.length > 1 ? item.product_title + item.variant_title : item.product_title,
      "sku": item.sku ? item.sku : "",
      "unitPrice": item.price,
      "shippingCost": 0,
      "shippingDescription": "",
      "unitTax": 0,
      "quantity": item.quantity
    };
  };

  function updateBreadItemsWithCurrency(items, currency) {
    for (i=0; i<items.length; i++) {
      var item = items[i];
      item.currency = currency;
      item.unitPrice = {currency: currency, value: item.unitPrice};
      item.shippingCost = {currency: currency, value: item.shippingCost};
      item.unitTax = {currency:currency, value: item.unitTax};

      items[i] = item;
    }

    return items;
  }

  function calculateTotalDiscount(cart) {
    var totalLineItemDiscount = 0;
    var totalCartLevelDiscount = 0;
    var items = cart.items;

    for (i=0; i<items.length; i++) {
      if ("line_level_total_discount" in items[i]) {
        totalLineItemDiscount += items[i].line_level_total_discount;
      }
    }

    if ("cart_level_discount_applications" in cart) {
      var cartLevelDiscounts = cart.cart_discount_applications;
      for (i=0; i<cartLevelDiscounts; i++) {
          totalCartLevelDiscount += cartLevelDiscounts[i].total_allocated_amount;
      }
    }

    return totalLineItemDiscount + totalCartLevelDiscount;
  }

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

  function configureProductButton() {
    var breadButtonElement = document.getElementById(getPDPDomId());

    getCart(function(err, cart) {
      if (err) {
        return console.log(ERROR_LABEL + 'query for shopify cart produced err: ', err);
      }

      var currency = cart.currency

      queryProductByHandleInUrl(function(product, err) {
        if (err) {
          return console.log(ERROR_LABEL + 'Query for product returned an error' + err);
        }
        
        if (!product) {
          console.log("Product not found");
          return;
        }
        
        if (product.variants.length == 1) {
          renderPDPButton(product, product.variants[0], currency, breadButtonElement);
          var vid = product.variants[0].id;
          return;
        }
    
        var variantIntervalId = setInterval(function() {
          var newVidEl = document.querySelectorAll("select[name='id'] :checked");
          if (newVidEl.length > 0) {
            var newVid = newVidEl[0].getAttribute('value');
            if (vid !== newVid) {
              vid = newVid;
              var variant = product.variants[variantIndexFromVid(product, vid)];
              renderPDPButton(product, variant, currency, breadButtonElement);
            }
          }
        }, 500);
      });

    });
    
  };

  function variantIndexFromVid(product, vid) {
    for (var x = 0; x < product.variants.length; x++) {
      if (product.variants[x].id == vid) return x;
    }
    return 0;
  };

  function renderPDPButton(product, variant, currency, breadButtonElement) {
    breadButtonElement.innerHTML = "";
    var setup = {integrationKey: integrationKey};
    window.RBCPayPlan.setup(setup);

    window.RBCPayPlan.on('INSTALLMENT:APPLICATION_DECISIONED', onApproved);
    window.RBCPayPlan.on('INSTALLMENT:APPLICATION_CHECKOUT', onCheckout); 
    
    var placement = {};
    placement.allowCheckout =  false;
    placement.domID = getPDPDomId();
    placement.order = {};
    
    placement.order.items = [mapProductToBreadItem(product, variant, currency)];
    placement.order.subTotal = {value: variant.price, currency: currency};
    placement.order.totalPrice = {value: variant.price, currency: currency};
    placement.order.currency = currency;
    placement.order.totalShipping = { value: 0, currency: currency }; //Not available in shopify product object
    placement.order.totalTax =  { value: 0, currency: currency }; //Not available in shopify product object
    placement.order.totalDiscounts =  { value: 0, currency: currency }; //Not available in shopify product object

    window.RBCPayPlan.registerPlacements([placement]);
    window.RBCPayPlan.__internal__.init();
  }

  function mapProductToBreadItem(product, variant, currency) {
    return {
      "name": [product.title, variant.title].join(", "),
      "sku": variant.sku ? variant.sku : "",
      "unitPrice": {currency: currency, value: variant.price},
      "shippingCost": {currency: currency, value: 0},
      "shippingDescription": "",
      "unitTax": {currency: currency, value: 0},
      "quantity": getQuantityValue(),
      "currency": currency
    };
  }

  function getQuantityValue() {
    // Default to quantity of 1
    var v = 1;
    var q = document.querySelector('input[name="quantity"]');

    if (q != null) {
      v = parseInt(q.value, 10);
    }
    return v;
  }

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

  
  function createButtons() {
    if ("RBCPayPlan" in window) {// SDK loaded
      if (document.querySelectorAll("#placement-cart").length > 0 || document.querySelectorAll("#bread-checkout-btn").length > 0) {
        configureCartButton();
      }

      if(document.querySelectorAll("#placement-pdp").length > 0 || document.querySelectorAll("#bread-checkout-btn-product").length > 0) {
        configureProductButton();
      }

    }else {
      console.log("SDK not loaded")
      //TODO: log error to datadog
    }
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

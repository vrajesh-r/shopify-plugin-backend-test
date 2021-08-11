const buildGateway = function() {
  let app = {};
  
  app.container = document.getElementById('app');


  app.getAllShops = function() {
    return fetch('/admin/data').then(r => r.json());
  };

  app.updateShopSettings = function(shop, settings) {
    const { enableAcceleratedCheckout, posAccess } = settings;
    return fetch('data/settings', {
      method: 'POST',
      mode: 'cors',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        shop,
        enableAcceleratedCheckout,
        posAccess
      })
    });
  };

  app.registerWebhooks = function(shop) {
    return fetch('webhooks', {
      method: 'POST',
      mode: 'cors',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        shop
      })
    });
  };



  const buildShopList = function(shops) {
    let ul = document.createElement('ul');
    shops.forEach(s => {
      let li = document.createElement('li');
      li.setAttribute('data-shop', s.shopName);
      li.setAttribute('data-enable', s.acceleratedCheckout);
      
      let h1 = document.createElement('h3');
      h1.innerHTML = s.shopName;
      
      let group1 = document.createElement('div');
      let label = document.createElement('label');
      label.innerHTML = 'Accelerated Checkout Enabled';
      let input = document.createElement('input');
      input.setAttribute('type', 'checkbox');
      input.classList.add('data-accelerated-checkout');
      input.checked = s.acceleratedCheckout;
      group1.appendChild(label);
      group1.appendChild(input);

      let group2 = document.createElement('div');
      let label2 = document.createElement('label');
      label2.innerHTML = 'POS Access';
      let input2 = document.createElement('input');
      input2.setAttribute('type', 'checkbox');
      input2.classList.add('data-pos-access');
      input2.checked = s.posAccess;
      group2.appendChild(label2);
      group2.appendChild(input2);
      
      li.appendChild(h1);
      li.appendChild(group1);
      li.appendChild(group2);

      let submit = document.createElement('button');
      submit.setAttribute('type', 'submit');
      submit.innerHTML = 'Submit';
      submit.addEventListener('click', e => {
        let shop = e.target.parentNode.getAttribute('data-shop');
        let enableAcceleratedCheckout = e.target.parentNode.querySelector('.data-accelerated-checkout').checked;
        let posAccess = e.target.parentNode.querySelector('.data-pos-access').checked;
        app.updateShopSettings(shop, {
          enableAcceleratedCheckout,
          posAccess
        })
        .then(d => {
          app.renderList();
        });
      });
      li.appendChild(submit);

      let webhookBtn = document.createElement('button');
      webhookBtn.setAttribute('type', 'submit');
      webhookBtn.innerHTML = 'Register Webhooks';
      webhookBtn.addEventListener('click', e => {
        let shop = e.target.parentNode.getAttribute('data-shop');
        app.registerWebhooks(shop)
        .then(d => {
          console.log(d.data);
          webhookBtn.innerHTML = "success";
        }).catch(err => {
          console.log(err);
          webhookBtn.innerHTML = "failed";
        });
      });
      li.appendChild(webhookBtn);

      ul.appendChild(li);
    });
    return ul;
  };

  app.renderList = function() {
    app.getAllShops()
    .then(d => {
      app.container.innerHTML = '';
      app.container.appendChild(buildShopList(d.shops));
    })
    .catch(e => {
      console.error(`Error: ${e}`);
    });
  }

  app.renderList();
};

document.addEventListener('DOMContentLoaded', buildGateway);

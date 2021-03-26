//---------------------------- DOM helpers -------------------------------------

// returns element
function byName(name) {
	return document.getElementsByName(name)[0];
}

// returns element
function byNameStr(name) {
	let element = byName(name);
	if(element) {
		return element.value;
	} else {
		return "";
	}
}

// returns element
function inElement(element, name) {
	return element.querySelector('[name="' + name + '"]');
}

// untrusted template literal input must be escaped
function htmlEscape(str) {
	return str.replace(/&/g, '&amp;') // must be first
		.replace(/>/g, '&gt;')
		.replace(/</g, '&lt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;')
		.replace(/`/g, '&#96;');
}

//---------------------------- number helpers ----------------------------------

// returns cents
function currency(str) {
	str = str.replace(",", "."); // what a crap
	return round(100.0 * parseFloat(str));
}

function centsToStr(cents) {
	return (cents / 100.0).toFixed(2).replace(".", ",") + " Euro";
}

// JavaScript is the worst. Math.round returns wrong results for negative input. Fixing that here.
function round(val) {
	var factor = 1;
	if(val < 0) {
		factor = -1;
		val = -val; // make positive
	}
	return factor * Math.round(val);
}

//---------------------------- template helpers --------------------------------

function prepareSubmit(event) {
	document.getElementById("data").value = JSON.stringify(readData());
}

function reloadCaptcha() {
	var e = document.getElementById("captcha-image");
	var q = "reload=" + (new Date()).getTime();
	var src  = e.src;
	var p = src.indexOf('?');
	if (p >= 0) {
		src = src.substr(0, p);
	}
	e.src = src + "?" + q
}

//---------------------------- drag and drop -----------------------------------

function allowDrop(ev) {
	ev.preventDefault();
}

function drag(ev) {
	// ev.target is a <tr>
	ev.dataTransfer.setData("text", ev.target.id);
}

function drop(ev) {
	var target = ev.target;
	// ev.target can be a <td>, <div>, etc.
	while(target.tagName != "TR") { // "the returned tag name is always in the canonical upper-case form"
		if(target.parentNode !== null) {
			target = target.parentNode;
		} else {
			return; // give up
		}
	}

	ev.preventDefault();
	var data = ev.dataTransfer.getData("text");
	target.parentNode.insertBefore(document.getElementById(data), target);

	updateView();
}

//---------------------------- edit form ---------------------------------------

var numTasks = 0;
const storeFeeBase = 1290;
const storeFeeShare = 0.02;

function addAddCost(taskNumber, data = {name: "", price: 0}, readOnly = false) {
	document.getElementById(`footer-${taskNumber}`).insertAdjacentHTML("afterend",
		`<tr name="add-cost">
			<td colspan="4"><input class="form-control" ${readOnly ? 'readonly' : ''} name="name"  value="${data['name']}"></td>
			<td><input class="form-control" ${readOnly ? 'readonly' : ''} name="price" value="${data['price'] / 100}" type="number" size="8" min="-10000" step="0.01" max="10000" onchange="updateView()"></td>
			<td>
				${readOnly ? '' : '<button type="button" class="btn btn-danger btn-sm" onclick="removeAddCost(this)">&ndash;</button>'}
			</td>
		</tr>`);
}

function addArticle(taskNumber, data = {link: "", properties: "", quantity: 1, price: 0}, readOnly = false) {

	var randomID = (Math.random().toString(36)+'00000000000000000').slice(2, 12); // required for drag and drop that tr has an id

	// draggable only if editable, else Firefox won't let us select the content of readonly inputs
	document.getElementById(`footer-${taskNumber}`).insertAdjacentHTML("beforebegin",
		`<tr name="article" id="${randomID}" ${readOnly ? '' : 'draggable="true" ondragstart="drag(event)" style="cursor: move; user-select: none;"'}>
			<td><input class="form-control" ${readOnly ? 'readonly' : ''} name="link"       value="${data['link']}"></td>
			<td><input class="form-control" ${readOnly ? 'readonly' : ''} name="properties" value="${data['properties']}"  placeholder="z. B. Größe und Farbe"></td>
			<td><input class="form-control" ${readOnly ? 'readonly' : ''} name="quantity"   value="${data['quantity']}"    type="number" size="3" min="1" step="1" onchange="updateView()"></td>
			<td><input class="form-control" ${readOnly ? 'readonly' : ''} name="price"      value="${data['price'] / 100}" type="number" size="8" min="0" step="0.01" max="10000" onchange="updateView()"></td>
			<td><div class="col-form-label" name="sum"></div></td>
			<td>
				${readOnly ? '' : '<button type="button" class="btn btn-danger btn-sm" onclick="removeArticle(this)">&ndash;</button>'}
			</td>
		</tr>`);
}

function addTask(data, readOnly = false, showHints = true, header = "") {

	var taskNumber = newTask(readOnly, showHints, header);
	for(const row of data["articles"]) {
		addArticle(taskNumber, row, readOnly);
	}

	document.querySelector(`[data-task="${taskNumber}"] [name="id"]`).value = data["id"];
	document.querySelector(`[data-task="${taskNumber}"] [name="merchant"]`).value = data["merchant"];
	document.querySelector(`[data-task="${taskNumber}"] [name="shipping-fee"]`).value = data["shipping-fee"] / 100;

	if(data["add-costs"]) {
		for(const addCost of data["add-costs"]) {
			addAddCost(taskNumber, addCost, readOnly);
		}
	}

	updateView();
}

function newTask(readOnly = false, showHints = true, header = "") {

	numTasks++;
	var taskNumber = numTasks;

	document.getElementById("add-task-paragraph").insertAdjacentHTML("beforebegin",
		`<div class="mt-3 bg-light" data-task="${taskNumber}" style="padding: 1em 1em 0 1em; border-top: solid #c0c0c0 1px; margin-bottom: 1em;">

	<h2>
		Einzelauftrag ${taskNumber}
		${readOnly ? '' : '<button type="button" class="btn btn-danger btn-sm" onclick="removeTask(this)">&ndash;</button>'}
	</h2>

	<input type="hidden" name="id">

	${header ? header : ''}

	<div class="form-group">
		<label>Link zum Versandhändler bzw. Anbieter</label>
		<input type="text" class="form-control" name="merchant" onchange="updateView()" ${readOnly ? 'readonly' : ''}>
		${showHints ? '<small class="form-text text-muted">Um Waren von verschiedenen Versandhändlern zu bestellen, füge bitte einen weiteren Einzelauftrag hinzu.</small>' : ''}
	</div>

	<table class="table">
		<thead>
			<tr>
				<th>Link oder Artikelnummer</th>
				<th>Eigenschaften</th>
				<th>Menge</th>
				<th>Einzelpreis</th>
				<th>Gesamtpreis</th>
				<th>
					${readOnly ? '' : '<button type="button" class="btn btn-success btn-sm" onclick="addArticle(' + taskNumber + ')">+</button>'}
				</th>
			</tr>
		</thead>
		<tbody ondrop="drop(event)" ondragover="allowDrop(event)">
			<tr id="footer-${taskNumber}">
				<td colspan="4" style="text-align: right">
					<div class="col-form-label">
						Kosten für Versand und Verpackung
					</div>
				</td>
				<td>
					<input class="form-control" name="shipping-fee" type="number" size="8" min="0.00" max="10000.00" step="0.01" onchange="updateView()" ${readOnly ? 'readonly' : ''}>
				</td>
				<td>
					${readOnly ? '' : '<button type="button" class="btn btn-success btn-sm" onclick="addAddCost(' + taskNumber + ')">+</button>'}
				</td>
			</tr>
			<tr>
				<td colspan="4" style="text-align: right"><strong>Bestellwert</strong></td>
				<td name="task-sum"></td>
				<td><!-- column with add and remove buttons --></td>
			</tr>
			<tr>
				<td colspan="4" style="text-align: right">+ Auftragsgebühr (12,90 € + 2 % des Bestellwerts)</td>
				<td name="store-fee"></td>
				<td><!-- column with add and remove buttons --></td>
			</tr>
			<tr>
				<td colspan="4" style="text-align: right"><strong>= Summe</strong></td>
				<td name="overall-sum"></td>
				<td><!-- column with add and remove buttons --></td>
			</tr>
		</tbody>
	</table>

</div>`);

	return taskNumber;
}

function removeAddCost(button) {
	var addCostElement = button.closest('[name="add-cost"]')
	addCostElement.parentNode.removeChild(addCostElement);
	updateView();
}

function removeArticle(button) {
	var articleElement = button.closest('[name="article"]')
	articleElement.parentNode.removeChild(articleElement);
	updateView();
}

function removeTask(button) {
	var taskElement = button.closest('[data-task]');
	taskElement.parentNode.removeChild(taskElement);
	updateView();
}

function readData(includeElements = false) {

	var data = {
		"client-name":                 byNameStr("client-name"),
		"client-contact":              byNameStr("client-contact"),
		"client-contact-protocol":     byNameStr("client-contact-protocol"),
		"shipping-address-supplement": byNameStr("shipping-address-supplement"),
		"shipping-first-name":         byNameStr("shipping-first-name"),
		"shipping-last-name":          byNameStr("shipping-last-name"),
		"shipping-postcode":           byNameStr("shipping-postcode"),
		"shipping-service":            byNameStr("shipping-service"),
		"shipping-street":             byNameStr("shipping-street"),
		"shipping-street-number":      byNameStr("shipping-street-number"),
		"shipping-town":               byNameStr("shipping-town"),
		"tasks":                       []
	};

	let dm = document.querySelector('input[name="delivery-method"]:checked');
	if(dm) {
		data["delivery-method"] = dm.value;
	}

	if(byName("reshipping-fee")) {
		data["reshipping-fee"] = currency(byName("reshipping-fee").value);
	}

	for(let taskElement of document.querySelectorAll('[data-task]')) {
		let task = {
			"id":           inElement(taskElement, "id").value,
			"merchant":     inElement(taskElement, "merchant").value,
			"shipping-fee": currency(inElement(taskElement, "shipping-fee").value),
			"articles":     [],
			"add-costs":    [],
			"element":      includeElements ? taskElement : null,
		};
		for(let articleElement of taskElement.querySelectorAll('[name="article"]')) {
			task["articles"].push({
				"link":       inElement(articleElement, "link").value,
				"price":      currency(inElement(articleElement, "price").value),
				"properties": inElement(articleElement, "properties").value,
				"quantity":   parseInt(inElement(articleElement, "quantity").value),
				"element":    includeElements ? articleElement : null,
			});
		}
		for(let addCostElement of taskElement.querySelectorAll('[name="add-cost"]')) {
			task["add-costs"].push({
				"name":  inElement(addCostElement, "name").value,
				"price": currency(inElement(addCostElement, "price").value) || 0 // prevent NaN if value is empty string
			});
		}
		data["tasks"].push(task);
	}

	return data;
}

function updateView() {

	var tbody = document.getElementById("summary-tbody");
	if(tbody) {
		tbody.textContent = ""; // faster than setting innerHTML
	}

	var data = readData(true); // money is in euro cents
	var totalSum = 0;
	var totalFee = 0;
	var totalSumWithFee = 0;

	for(let task of data["tasks"]) {

		var taskSum = 0; // integer

		for(const article of task["articles"]) {
			var result = "";
			var quantity = article["quantity"];
			var price = article["price"];
			if(quantity > 0 && price > 0) {
				var rowSum = quantity * price;
				taskSum += rowSum;
				result = centsToStr(rowSum);
			}
			inElement(article["element"], "sum").innerHTML = result;
		}

		taskSum += task["shipping-fee"] || 0;

		for(const addCost of task["add-costs"]) {
			taskSum += addCost["price"];
		}

		var taskFee = round(storeFeeShare * taskSum) + storeFeeBase;
		var taskSumWithFee = taskSum + taskFee;

		inElement(task["element"], "task-sum").innerHTML = centsToStr(taskSum);
		inElement(task["element"], "store-fee").innerHTML = centsToStr(taskFee);
		inElement(task["element"], "overall-sum").innerHTML = centsToStr(taskSumWithFee);

		if(tbody) {
			tbody.insertAdjacentHTML("beforeend",
				`<tr>
					<td>Auftrag ${htmlEscape(task["element"].dataset.task)}: ${htmlEscape(task["merchant"])}</td>
					<td>${centsToStr(taskSum)}</td>
					<td>${centsToStr(taskFee)}</td>
					<td>${centsToStr(taskSumWithFee)}</td>
				</tr>`
			); // take task number from element because it is not returned by readData (because it does not contribute to )
		}

		totalSum += taskSum
		totalFee += taskFee
		totalSumWithFee += taskSumWithFee;
	}

	let shippingElem = document.getElementById("delivery-method-shipping");
	if(shippingElem && shippingElem.checked == true) {

		// 1. default value
		var reshippingFee = parseInt(document.querySelector('select[name="shipping-service"] option:checked').dataset.minCost);

		// 2. data value
		if(data["reshipping-fee"]) {
			reshippingFee = data["reshipping-fee"];
		}

		// 3. entered value, if it is greater
		if(byName("reshipping-fee")) {
			var cents = currency(byName("reshipping-fee").value);
			if(reshippingFee < cents) {
				reshippingFee = cents;
			}
		}

		var reshippingFeeStr = centsToStr(reshippingFee);

		if(tbody) {
			tbody.insertAdjacentHTML("beforeend",
				`<tr>
					<td>Weiterversand</td>
					<td>${reshippingFeeStr}</td>
				</tr>`
			);
		}

		totalSumWithFee += reshippingFee;
	}

	if(tbody) {
		tbody.insertAdjacentHTML("beforeend",
			`<tr>
				<td><strong>Gesamtsumme</td>
				<td><strong>${centsToStr(totalSum)}</strong></td>
				<td><strong>${centsToStr(totalFee)}</strong></td>
				<td><strong>${centsToStr(totalSumWithFee)}</strong></td>
			</tr>`
		);
	}
}

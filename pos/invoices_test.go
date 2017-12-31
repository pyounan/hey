package pos

import (
	"bytes"
	"encoding/json"
	"pos-proxy/pos/models"
	"testing"
)

func TestHandleOperaPayments(t *testing.T) {
	invoice := models.Invoice{}
	buf := bytes.NewBufferString(`{
    "id": null,
    "invoice_number": "4-122",
    "posinvoicelineitem_set": [
      {
        "item": 2,
        "qty": 1,
        "submitted_qty": 1,
        "returned_qty": 0,
        "description": "HOT DOG",
        "comment": "",
        "unit_price": 70,
        "price": 70,
        "net_amount": 63.63636363636363,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "lineitem_type": "",
        "is_condiment": false,
        "condimentlineitem_set": [
          {
            "condiment": 6,
            "posinvoicelineitem": 0,
            "name": "Olive",
            "price": 10,
            "net_amount": 10,
            "tax_amount": 0,
            "vat_code": "D",
            "vat_percentage": 0,
            "attached_attributes": {
              "expense_account": 94,
              "house_use_expense_account": 95,
              "revenue_department": 35
            },
            "storemenuitemconfig": 46
          }
        ],
        "itemcondimentgroup_set": [
          {
            "condiment_group": 1,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c68146629a",
                "group": 1,
                "id": 1,
                "item": null,
                "name": "well done",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629b",
                "group": 1,
                "id": 2,
                "item": null,
                "name": "rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629c",
                "group": 1,
                "id": 3,
                "item": null,
                "name": "medium",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629d",
                "group": 1,
                "id": 4,
                "item": null,
                "name": "medium rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c6814662b8",
                "group": 1,
                "id": 5,
                "item": null,
                "name": "charred",
                "price": 0
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Cooking"
          },
          {
            "condiment_group": 2,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c6814662b6",
                "group": 2,
                "id": 6,
                "item": 55,
                "name": "Olive",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b7",
                "group": 2,
                "id": 7,
                "item": 56,
                "name": "cheese",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b9",
                "group": 2,
                "id": 8,
                "item": 2,
                "name": "Hot Dog",
                "price": 40
              },
              {
                "_id": "59d4e8bf566d43c6814662ba",
                "group": 2,
                "id": 9,
                "item": 11,
                "name": "Cola Drink",
                "price": 5
              },
              {
                "_id": "59d4e8bf566d43c6814662bb",
                "group": 2,
                "id": 10,
                "item": 12,
                "name": "7up Drink",
                "price": 5
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Extras"
          }
        ],
        "is_discount": false,
        "applied_discounts": [],
        "grouped_applieddiscounts": [],
        "attached_attributes": {
          "color": "#ff0000",
          "course": {
            "id": 3
          },
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 28,
          "waste_department": 28
        },
        "course": 3,
        "storemenuitemconfig": 2,
        "open_item": false,
        "open_price": false,
        "returned_ids": null,
        "frontend_id": "f05e17b7-1eda-38e5-cf03-bfa87039f627",
        "updated_on": "",
        "store_unit": 2,
        "base_unit": "Each",
        "original_frontend_id": null,
        "original_line_item_id": null,
        "posinvoice": null,
        "index": 0
      },
      {
        "item": 2,
        "qty": 9,
        "submitted_qty": 9,
        "returned_qty": 0,
        "description": "HOT DOG",
        "comment": "",
        "unit_price": 70,
        "price": 630,
        "net_amount": 572.7272727272726,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "lineitem_type": "",
        "is_condiment": false,
        "condimentlineitem_set": [
          {
            "condiment": 8,
            "posinvoicelineitem": 0,
            "name": "Hot Dog",
            "price": 40,
            "net_amount": 327.27272727272725,
            "tax_amount": 0,
            "vat_code": "D",
            "vat_percentage": 0,
            "attached_attributes": {
              "color": "#ff0000",
              "course": {
                "id": 3
              },
              "expense_account": 94,
              "house_use_expense_account": 95,
              "revenue_department": 28,
              "waste_department": 28
            },
            "storemenuitemconfig": 2
          }
        ],
        "itemcondimentgroup_set": [
          {
            "condiment_group": 1,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c68146629a",
                "group": 1,
                "id": 1,
                "item": null,
                "name": "well done",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629b",
                "group": 1,
                "id": 2,
                "item": null,
                "name": "rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629c",
                "group": 1,
                "id": 3,
                "item": null,
                "name": "medium",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629d",
                "group": 1,
                "id": 4,
                "item": null,
                "name": "medium rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c6814662b8",
                "group": 1,
                "id": 5,
                "item": null,
                "name": "charred",
                "price": 0
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Cooking"
          },
          {
            "condiment_group": 2,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c6814662b6",
                "group": 2,
                "id": 6,
                "item": 55,
                "name": "Olive",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b7",
                "group": 2,
                "id": 7,
                "item": 56,
                "name": "cheese",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b9",
                "group": 2,
                "id": 8,
                "item": 2,
                "name": "Hot Dog",
                "price": 40
              },
              {
                "_id": "59d4e8bf566d43c6814662ba",
                "group": 2,
                "id": 9,
                "item": 11,
                "name": "Cola Drink",
                "price": 5
              },
              {
                "_id": "59d4e8bf566d43c6814662bb",
                "group": 2,
                "id": 10,
                "item": 12,
                "name": "7up Drink",
                "price": 5
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Extras"
          }
        ],
        "is_discount": false,
        "applied_discounts": [],
        "grouped_applieddiscounts": [],
        "attached_attributes": {
          "color": "#ff0000",
          "course": {
            "id": 3
          },
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 28,
          "waste_department": 28
        },
        "course": 3,
        "storemenuitemconfig": 2,
        "open_item": false,
        "open_price": false,
        "returned_ids": null,
        "frontend_id": "d1e181b2-8de6-2c1d-eaa2-53312c1bd16d",
        "updated_on": "",
        "store_unit": 2,
        "base_unit": "Each",
        "original_frontend_id": null,
        "original_line_item_id": null,
        "posinvoice": null,
        "index": 1,
        "lastchildincourse": true
      }
    ],
    "grouped_lineitems": [
      {
        "item": 2,
        "qty": 1,
        "submitted_qty": 1,
        "returned_qty": 0,
        "description": "HOT DOG",
        "comment": "",
        "unit_price": 70,
        "price": 70,
        "net_amount": 70,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "lineitem_type": "",
        "is_condiment": false,
        "condimentlineitem_set": [
          {
            "condiment": 6,
            "posinvoicelineitem": 0,
            "name": "Olive",
            "price": 10,
            "net_amount": 10,
            "tax_amount": 0,
            "vat_code": "D",
            "vat_percentage": 0,
            "attached_attributes": {
              "expense_account": 94,
              "house_use_expense_account": 95,
              "revenue_department": 35
            },
            "storemenuitemconfig": 46,
            "description": "Olive",
            "qty": 1,
            "is_condiment": true,
            "is_discount": false,
            "item": null
          }
        ],
        "itemcondimentgroup_set": [
          {
            "condiment_group": 1,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c68146629a",
                "group": 1,
                "id": 1,
                "item": null,
                "name": "well done",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629b",
                "group": 1,
                "id": 2,
                "item": null,
                "name": "rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629c",
                "group": 1,
                "id": 3,
                "item": null,
                "name": "medium",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629d",
                "group": 1,
                "id": 4,
                "item": null,
                "name": "medium rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c6814662b8",
                "group": 1,
                "id": 5,
                "item": null,
                "name": "charred",
                "price": 0
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Cooking"
          },
          {
            "condiment_group": 2,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c6814662b6",
                "group": 2,
                "id": 6,
                "item": 55,
                "name": "Olive",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b7",
                "group": 2,
                "id": 7,
                "item": 56,
                "name": "cheese",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b9",
                "group": 2,
                "id": 8,
                "item": 2,
                "name": "Hot Dog",
                "price": 40
              },
              {
                "_id": "59d4e8bf566d43c6814662ba",
                "group": 2,
                "id": 9,
                "item": 11,
                "name": "Cola Drink",
                "price": 5
              },
              {
                "_id": "59d4e8bf566d43c6814662bb",
                "group": 2,
                "id": 10,
                "item": 12,
                "name": "7up Drink",
                "price": 5
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Extras"
          }
        ],
        "is_discount": false,
        "applied_discounts": [],
        "grouped_applieddiscounts": [],
        "attached_attributes": {
          "color": "#ff0000",
          "course": {
            "id": 3
          },
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 28,
          "waste_department": 28
        },
        "course": 3,
        "storemenuitemconfig": 2,
        "open_item": false,
        "open_price": false,
        "returned_ids": null,
        "frontend_id": "f05e17b7-1eda-38e5-cf03-bfa87039f627",
        "updated_on": "",
        "store_unit": 2,
        "base_unit": "Each",
        "original_frontend_id": null,
        "original_line_item_id": null,
        "posinvoice": null,
        "index": 0
      },
      {
        "condiment": 6,
        "posinvoicelineitem": 0,
        "name": "Olive",
        "price": 10,
        "net_amount": 10,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "attached_attributes": {
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 35
        },
        "storemenuitemconfig": 46,
        "description": "Olive",
        "qty": 1,
        "is_condiment": true,
        "is_discount": false,
        "item": null
      },
      {
        "item": 2,
        "qty": 9,
        "submitted_qty": 9,
        "returned_qty": 0,
        "description": "HOT DOG",
        "comment": "",
        "unit_price": 70,
        "price": 630,
        "net_amount": 630,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "lineitem_type": "",
        "is_condiment": false,
        "condimentlineitem_set": [
          {
            "condiment": 8,
            "posinvoicelineitem": 0,
            "name": "Hot Dog",
            "price": 360,
            "net_amount": 360,
            "tax_amount": 0,
            "vat_code": "D",
            "vat_percentage": 0,
            "attached_attributes": {
              "color": "#ff0000",
              "course": {
                "id": 3
              },
              "expense_account": 94,
              "house_use_expense_account": 95,
              "revenue_department": 28,
              "waste_department": 28
            },
            "storemenuitemconfig": 2,
            "description": "Hot Dog",
            "qty": 9,
            "is_condiment": true,
            "is_discount": false,
            "item": null
          }
        ],
        "itemcondimentgroup_set": [
          {
            "condiment_group": 1,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c68146629a",
                "group": 1,
                "id": 1,
                "item": null,
                "name": "well done",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629b",
                "group": 1,
                "id": 2,
                "item": null,
                "name": "rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629c",
                "group": 1,
                "id": 3,
                "item": null,
                "name": "medium",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629d",
                "group": 1,
                "id": 4,
                "item": null,
                "name": "medium rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c6814662b8",
                "group": 1,
                "id": 5,
                "item": null,
                "name": "charred",
                "price": 0
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Cooking"
          },
          {
            "condiment_group": 2,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c6814662b6",
                "group": 2,
                "id": 6,
                "item": 55,
                "name": "Olive",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b7",
                "group": 2,
                "id": 7,
                "item": 56,
                "name": "cheese",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b9",
                "group": 2,
                "id": 8,
                "item": 2,
                "name": "Hot Dog",
                "price": 40
              },
              {
                "_id": "59d4e8bf566d43c6814662ba",
                "group": 2,
                "id": 9,
                "item": 11,
                "name": "Cola Drink",
                "price": 5
              },
              {
                "_id": "59d4e8bf566d43c6814662bb",
                "group": 2,
                "id": 10,
                "item": 12,
                "name": "7up Drink",
                "price": 5
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Extras"
          }
        ],
        "is_discount": false,
        "applied_discounts": [],
        "grouped_applieddiscounts": [],
        "attached_attributes": {
          "color": "#ff0000",
          "course": {
            "id": 3
          },
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 28,
          "waste_department": 28
        },
        "course": 3,
        "storemenuitemconfig": 2,
        "open_item": false,
        "open_price": false,
        "returned_ids": null,
        "frontend_id": "d1e181b2-8de6-2c1d-eaa2-53312c1bd16d",
        "updated_on": "",
        "store_unit": 2,
        "base_unit": "Each",
        "original_frontend_id": null,
        "original_line_item_id": null,
        "posinvoice": null,
        "index": 1,
        "lastchildincourse": true
      },
      {
        "condiment": 8,
        "posinvoicelineitem": 0,
        "name": "Hot Dog",
        "price": 360,
        "net_amount": 360,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "attached_attributes": {
          "color": "#ff0000",
          "course": {
            "id": 3
          },
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 28,
          "waste_department": 28
        },
        "storemenuitemconfig": 2,
        "description": "Hot Dog",
        "qty": 9,
        "is_condiment": true,
        "is_discount": false,
        "item": null
      }
    ],
    "table": 1,
    "events": [],
    "audit_date": "2017-09-10",
    "cashier": 2,
    "cashier_details": "3/pos_testmt",
    "cashier_number": 3,
    "created_on": "2017-10-04T14:47:26.465Z",
    "frontend_id": "35a578eb-88f5-a7cc-d158-a64c969d9e46",
    "is_settled": false,
    "paid_amount": 0,
    "pax": 1,
    "walkin_name": null,
    "profile_name": null,
    "profile_details": null,
    "store": 2,
    "store_description": "Pizzeria",
    "subtotal": 750,
    "table_number": 1,
    "takeout": false,
    "terminal_id": 4,
    "terminal_description": "",
    "total": 1070,
    "fdm_responses": null,
    "pospayment": null,
    "room_details": "",
    "house_use": false,
    "print_count": 0,
    "taxes": {},
    "ordered_posinvoicelineitem_set": [
      {
        "item": 2,
        "qty": 1,
        "submitted_qty": 1,
        "returned_qty": 0,
        "description": "HOT DOG",
        "comment": "",
        "unit_price": 70,
        "price": 70,
        "net_amount": 63.63636363636363,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "lineitem_type": "",
        "is_condiment": false,
        "condimentlineitem_set": [
          {
            "condiment": 6,
            "posinvoicelineitem": 0,
            "name": "Olive",
            "price": 10,
            "net_amount": 10,
            "tax_amount": 0,
            "vat_code": "D",
            "vat_percentage": 0,
            "attached_attributes": {
              "expense_account": 94,
              "house_use_expense_account": 95,
              "revenue_department": 35
            },
            "storemenuitemconfig": 46
          }
        ],
        "itemcondimentgroup_set": [
          {
            "condiment_group": 1,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c68146629a",
                "group": 1,
                "id": 1,
                "item": null,
                "name": "well done",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629b",
                "group": 1,
                "id": 2,
                "item": null,
                "name": "rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629c",
                "group": 1,
                "id": 3,
                "item": null,
                "name": "medium",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629d",
                "group": 1,
                "id": 4,
                "item": null,
                "name": "medium rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c6814662b8",
                "group": 1,
                "id": 5,
                "item": null,
                "name": "charred",
                "price": 0
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Cooking"
          },
          {
            "condiment_group": 2,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c6814662b6",
                "group": 2,
                "id": 6,
                "item": 55,
                "name": "Olive",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b7",
                "group": 2,
                "id": 7,
                "item": 56,
                "name": "cheese",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b9",
                "group": 2,
                "id": 8,
                "item": 2,
                "name": "Hot Dog",
                "price": 40
              },
              {
                "_id": "59d4e8bf566d43c6814662ba",
                "group": 2,
                "id": 9,
                "item": 11,
                "name": "Cola Drink",
                "price": 5
              },
              {
                "_id": "59d4e8bf566d43c6814662bb",
                "group": 2,
                "id": 10,
                "item": 12,
                "name": "7up Drink",
                "price": 5
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Extras"
          }
        ],
        "is_discount": false,
        "applied_discounts": [],
        "grouped_applieddiscounts": [],
        "attached_attributes": {
          "color": "#ff0000",
          "course": {
            "id": 3
          },
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 28,
          "waste_department": 28
        },
        "course": 3,
        "storemenuitemconfig": 2,
        "open_item": false,
        "open_price": false,
        "returned_ids": null,
        "frontend_id": "f05e17b7-1eda-38e5-cf03-bfa87039f627",
        "updated_on": "",
        "store_unit": 2,
        "base_unit": "Each",
        "original_frontend_id": null,
        "original_line_item_id": null,
        "posinvoice": null,
        "index": 0
      },
      {
        "item": 2,
        "qty": 9,
        "submitted_qty": 9,
        "returned_qty": 0,
        "description": "HOT DOG",
        "comment": "",
        "unit_price": 70,
        "price": 630,
        "net_amount": 572.7272727272726,
        "tax_amount": 0,
        "vat_code": "D",
        "vat_percentage": 0,
        "lineitem_type": "",
        "is_condiment": false,
        "condimentlineitem_set": [
          {
            "condiment": 8,
            "posinvoicelineitem": 0,
            "name": "Hot Dog",
            "price": 40,
            "net_amount": 327.27272727272725,
            "tax_amount": 0,
            "vat_code": "D",
            "vat_percentage": 0,
            "attached_attributes": {
              "color": "#ff0000",
              "course": {
                "id": 3
              },
              "expense_account": 94,
              "house_use_expense_account": 95,
              "revenue_department": 28,
              "waste_department": 28
            },
            "storemenuitemconfig": 2
          }
        ],
        "itemcondimentgroup_set": [
          {
            "condiment_group": 1,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c68146629a",
                "group": 1,
                "id": 1,
                "item": null,
                "name": "well done",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629b",
                "group": 1,
                "id": 2,
                "item": null,
                "name": "rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629c",
                "group": 1,
                "id": 3,
                "item": null,
                "name": "medium",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c68146629d",
                "group": 1,
                "id": 4,
                "item": null,
                "name": "medium rare",
                "price": 0
              },
              {
                "_id": "59d4e8bf566d43c6814662b8",
                "group": 1,
                "id": 5,
                "item": null,
                "name": "charred",
                "price": 0
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Cooking"
          },
          {
            "condiment_group": 2,
            "condiments": [
              {
                "_id": "59d4e8bf566d43c6814662b6",
                "group": 2,
                "id": 6,
                "item": 55,
                "name": "Olive",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b7",
                "group": 2,
                "id": 7,
                "item": 56,
                "name": "cheese",
                "price": 10
              },
              {
                "_id": "59d4e8bf566d43c6814662b9",
                "group": 2,
                "id": 8,
                "item": 2,
                "name": "Hot Dog",
                "price": 40
              },
              {
                "_id": "59d4e8bf566d43c6814662ba",
                "group": 2,
                "id": 9,
                "item": 11,
                "name": "Cola Drink",
                "price": 5
              },
              {
                "_id": "59d4e8bf566d43c6814662bb",
                "group": 2,
                "id": 10,
                "item": 12,
                "name": "7up Drink",
                "price": 5
              }
            ],
            "max": 1,
            "min": 0,
            "name": "Extras"
          }
        ],
        "is_discount": false,
        "applied_discounts": [],
        "grouped_applieddiscounts": [],
        "attached_attributes": {
          "color": "#ff0000",
          "course": {
            "id": 3
          },
          "expense_account": 94,
          "house_use_expense_account": 95,
          "revenue_department": 28,
          "waste_department": 28
        },
        "course": 3,
        "storemenuitemconfig": 2,
        "open_item": false,
        "open_price": false,
        "returned_ids": null,
        "frontend_id": "d1e181b2-8de6-2c1d-eaa2-53312c1bd16d",
        "updated_on": "",
        "store_unit": 2,
        "base_unit": "Each",
        "original_frontend_id": null,
        "original_line_item_id": null,
        "posinvoice": null,
        "index": 1,
        "lastchildincourse": true
      }
    ],
    "last_payment_date": "2017-09-10"
  }`)
	if err := json.NewDecoder(buf).Decode(&invoice); err != nil {
		t.Error(err)
		t.Fail()
	}
	HandleOperaPayments(invoice, 1)
}

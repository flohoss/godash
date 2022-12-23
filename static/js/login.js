import { createApp, nextTick } from "https://unpkg.com/petite-vue?module";

function FieldComponent(props) {
  return {
    $template: "#field-component-template",
    id: props.id,
    field: props.field,
    get invalidMessage() {
      return props.invalidMessage();
    },
  };
}

function requiredFieldMessage(what) {
  return what + " is a required field";
}

createApp({
  FieldComponent,
  $delimiters: ["[[", "]]"],
  submitted: false,
  invalids: {},
  fields: {
    name: {
      label: "Name",
      value: "",
      validation: { message: requiredFieldMessage("Name"), test: (value) => value },
    },
    password: {
      label: "Password",
      value: "",
      validation: { message: requiredFieldMessage("Password"), test: (value) => value },
    },
  },
  get isInvalid() {
    return !!Object.values(this.invalids).filter((key) => key).length;
  },
  validate() {
    this.invalids = {};
    Object.entries(this.fields).forEach((key) => {
      this.validateField(key[0], key[1]);
    });
  },
  validateField(fieldKey, field) {
    this.invalids[fieldKey] = false;
    if (!field.validation.test(field.value)) {
      this.invalids[fieldKey] = field.validation.message;
    }
  },
  submit() {
    this.validate();
    if (this.isInvalid) return;
    console.log("doing submit", this.fields);
    this.submitted = true;
  },
  handleInput(event) {
    if (event.key === "Escape") {
      this.fields[event.target.id].value = "";
    } else if (event.key === "Enter") {
      this.submit();
    }
  },
}).mount();

package variants

import (
	"fmt"
	"strings"
)

// Hotel is variant for the Jezero game
type Hotel struct{}

// NewHotel creates a new instance of Hotel
func NewHotel() *Hotel { return &Hotel{} }

////////////////////////////////////////////////////////////////////////////////
//revive:disable:exported

func (h *Hotel) ViewTemplateName() string    { return "viewHotel" }
func (h *Hotel) TemplateStateName() string   { return "Bojlery" }
func (h *Hotel) TemplateStateSymbol() string { return "°F" }
func (h *Hotel) TemplateMoneyName() string   { return "Spokojenost" }
func (h *Hotel) TemplateMoneySymbol() string { return "" }

func (h *Hotel) NopName() string        { return "Zůstat špinavý" }
func (h *Hotel) EcoName() string        { return "Umýt se rychle a úsporně" }
func (h *Hotel) HarvestName() string    { return "Naložit se do horké lázně" }
func (h *Hotel) CleaningName() string   { return "Přiložit pod kotlem" }
func (h *Hotel) InspectionName() string { return "Zavolat uklízečku na kontrolu" }
func (h *Hotel) EspionageName() string  { return "Nastražit kameru do sprch" }

////////////////////////////////////////////////////////////////////////////////

func (h *Hotel) EcoMessage(money int, pollution int) string {
	return fmt.Sprintf("Umyli jste se rychle a úsporně ve vlažné vodě. Vaše spokojenost se zvýšila o %d a váš bojler tím zchladl o %d°F.", money, pollution)
}
func (h *Hotel) HarvestPenaltyMessage(penalty int) string {
	return fmt.Sprintf("Vaše horká lázeň byla odhalena uklízečkou! Vaše spokojenost klesla o %d.", penalty)
}
func (h *Hotel) HarvestSuccessMessage(money int, pollution int) string {
	return fmt.Sprintf("Naložili jste se do horké lázně a spotřebovali jste spoustu teplé vody. Váš bojler tím vychladl o %d°F, ale vaše spokojenost vzrostla o %d.", pollution, money)
}
func (h *Hotel) CleaningMessage(cleaning int) string {
	return fmt.Sprintf("Přiložili jste pod kotlem a tím ohřáli váš bojler o %d°F.", cleaning)
}
func (h *Hotel) InspectionMessage() string {
	return "Požádali jste uklízečku u kontrolu všech sprch. Pokud někdo v minulém kole prováděl něco špatného, tak byl potrestán."
}
func (h *Hotel) EspionageFailMessage() string {
	return "Nastražené kamery bohužel zabavila uklízečka na kontrole, nic jste nezjistili."
}
func (h *Hotel) EspionageSuccessMessage(teamActions map[string]string) string {
	results := []string{}
	for team, action := range teamActions {
		results = append(results, fmt.Sprintf("<li>%s: <b>%s</b></li>", team, action))
	}
	return fmt.Sprintf("Úspěšně jste nastražili kamery do sprch. Akce v proběhlém kole:<ul>\n%s\n</ul>", strings.Join(results, "\n"))
}

func (h *Hotel) GlobalMessage(reduce string, increase string, change int) string {
	return fmt.Sprintf("<b>Sálání:</b> Sáláním se bojler %s ochladil o %d°F a bojler %s se naopak ohřál o %d°F.", reduce, change, increase, change)
}
